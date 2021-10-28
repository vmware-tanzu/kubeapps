import { JSONSchemaType } from "ajv";
import {
  AvailablePackageDetail,
  Context,
  InstalledPackageDetail,
  InstalledPackageReference,
  InstalledPackageSummary,
  VersionReference,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { ThunkAction } from "redux-thunk";
import PackagesService from "shared/PackagesService";
import {
  CreateError,
  DeleteError,
  FetchError,
  FetchWarning,
  IStoreState,
  RollbackError,
  UnprocessableEntity,
  UpgradeError,
} from "shared/types";
import { PluginNames } from "shared/utils";
import { ActionType, deprecated } from "typesafe-actions";
import { App } from "../shared/App";
import { validate } from "../shared/schema";

const { createAction } = deprecated;

export const requestApps = createAction("REQUEST_APPS");

export const listApps = createAction("REQUEST_APP_LIST");

export const receiveAppList = createAction("RECEIVE_APP_LIST", resolve => {
  return (apps: InstalledPackageSummary[]) => resolve(apps);
});

export const requestDeleteInstalledPackage = createAction("REQUEST_DELETE_INSTALLED_PACKAGE");

export const receiveDeleteInstalledPackage = createAction(
  "RECEIVE_DELETE_INSTALLED_PACKAGE_CONFIRMATION",
);

export const requestInstallPackage = createAction("REQUEST_INSTALL_PACKAGE");

export const receiveInstallPackage = createAction("RECEIVE_INSTALL_PACKAGE_CONFIRMATION");

export const requestUpdateInstalledPackage = createAction("REQUEST_UPDATE_INSTALLED_PACKAGE");

export const receiveUpdateInstalledPackage = createAction(
  "RECEIVE_UPDATE_INSTALLED_PACKAGE_CONFIRMATION",
);

export const requestRollbackInstalledPackage = createAction("REQUEST_ROLLBACK_INSTALLED_PACKAGE");

export const receiveRollbackInstalledPackage = createAction(
  "RECEIVE_ROLLBACK_INSTALLED_PACKAGE_CONFIRMATION",
);

export const errorApp = createAction("ERROR_APP", resolve => {
  return (err: FetchError | CreateError | UpgradeError | RollbackError | DeleteError) =>
    resolve(err);
});

export const selectApp = createAction("SELECT_APP", resolve => {
  // TODO(agamez): remove it once we return the generated resources as part of the InstalledPackageDetail.
  return (app: InstalledPackageDetail, manifest: any, details?: AvailablePackageDetail) =>
    resolve({ app, manifest, details });
});

const allActions = [
  listApps,
  requestApps,
  receiveAppList,
  requestDeleteInstalledPackage,
  receiveDeleteInstalledPackage,
  requestInstallPackage,
  receiveInstallPackage,
  requestUpdateInstalledPackage,
  receiveUpdateInstalledPackage,
  requestRollbackInstalledPackage,
  receiveRollbackInstalledPackage,
  errorApp,
  selectApp,
];

export type AppsAction = ActionType<typeof allActions[number]>;

export function getApp(
  installedPackageRef?: InstalledPackageReference,
): ThunkAction<Promise<void>, IStoreState, null, AppsAction> {
  return async dispatch => {
    dispatch(requestApps());
    try {
      // TODO(agamez/minelson): remove it once we enable the getting resources for
      // an installed package in the API.
      // TODO(minelson): Also remove conditional behaviour once resources can be
      // fetched in both flux and helm plugins.
      const legacyResponse =
        installedPackageRef?.plugin?.name === PluginNames.PACKAGES_HELM
          ? await App.getRelease(installedPackageRef)
          : undefined;
      // Get the details of an installed package
      const { installedPackageDetail } = await App.GetInstalledPackageDetail(installedPackageRef);

      // Not all plugins (flux) are cluster aware.
      if (
        installedPackageDetail &&
        !installedPackageDetail?.installedPackageRef?.context?.cluster
      ) {
        installedPackageDetail.installedPackageRef = installedPackageRef;
      }
      // For local packages with no references to any available packages (eg.a local package for development)
      // we aren't able to get the details, but still want to display the available data so far
      let availablePackageDetail;
      try {
        // Get the details of the available package that corresponds to the installed package
        const resp = await PackagesService.getAvailablePackageDetail(
          installedPackageDetail?.availablePackageRef,
          installedPackageDetail?.currentVersion?.pkgVersion,
        );
        availablePackageDetail = resp.availablePackageDetail;
      } catch (e: any) {
        dispatch(
          errorApp(
            new FetchWarning(
              "this package has missing information, some actions might not be available.",
            ),
          ),
        );
      }
      dispatch(
        selectApp(installedPackageDetail!, legacyResponse?.manifest, availablePackageDetail),
      );
    } catch (e: any) {
      dispatch(errorApp(new FetchError(e.message)));
    }
  };
}

export function deleteInstalledPackage(
  installedPackageRef: InstalledPackageReference,
): ThunkAction<Promise<boolean>, IStoreState, null, AppsAction> {
  return async dispatch => {
    dispatch(requestDeleteInstalledPackage());
    try {
      await App.DeleteInstalledPackage(installedPackageRef);
      dispatch(receiveDeleteInstalledPackage());
      return true;
    } catch (e: any) {
      dispatch(errorApp(new DeleteError(e.message)));
      return false;
    }
  };
}

// fetchApps returns a list of apps for other actions to compose on top of it
export function fetchApps(
  cluster: string,
  namespace?: string,
): ThunkAction<Promise<InstalledPackageSummary[]>, IStoreState, null, AppsAction> {
  return async dispatch => {
    dispatch(listApps());
    let installedPackageSummaries: InstalledPackageSummary[];
    try {
      const res = await App.GetInstalledPackageSummaries(cluster, namespace);
      installedPackageSummaries = res?.installedPackageSummaries;

      // Some plugins are not cluster aware, so initialize the cluster.
      // TODO(minelson) Let's just ensure all plugins send the cluster even if
      // they don't support multicluster?
      installedPackageSummaries = installedPackageSummaries.map(pkg => {
        pkg.installedPackageRef!.context!.cluster = cluster;
        return pkg;
      });

      dispatch(receiveAppList(installedPackageSummaries));
      return installedPackageSummaries;
    } catch (e: any) {
      dispatch(errorApp(new FetchError(e.message)));
      return [];
    }
  };
}

export function installPackage(
  targetCluster: string,
  targetNamespace: string,
  availablePackageDetail: AvailablePackageDetail,
  releaseName: string,
  values?: string,
  schema?: JSONSchemaType<any>,
): ThunkAction<Promise<boolean>, IStoreState, null, AppsAction> {
  return async dispatch => {
    dispatch(requestInstallPackage());
    try {
      if (values && schema) {
        const validation = validate(values, schema);
        if (!validation.valid) {
          const errorText =
            validation.errors &&
            validation.errors.map(e => `  - ${e.instancePath}: ${e.message}`).join("\n");
          throw new UnprocessableEntity(
            `The given values don't match the required format. The following errors were found:\n${errorText}`,
          );
        }
      }
      if (
        availablePackageDetail?.availablePackageRef &&
        availablePackageDetail?.version?.pkgVersion
      ) {
        await App.CreateInstalledPackage(
          { cluster: targetCluster, namespace: targetNamespace } as Context,
          releaseName,
          availablePackageDetail.availablePackageRef,
          { version: availablePackageDetail.version.pkgVersion } as VersionReference,
          values,
        );
        dispatch(receiveInstallPackage());
        return true;
      } else {
        dispatch(
          errorApp(
            new CreateError("This package does not contain enough information to be installed"),
          ),
        );
        return false;
      }
    } catch (e: any) {
      dispatch(errorApp(new CreateError(e.message)));
      return false;
    }
  };
}

export function updateInstalledPackage(
  installedPackageRef: InstalledPackageReference,
  availablePackageDetail: AvailablePackageDetail,
  values?: string,
  schema?: JSONSchemaType<any>,
): ThunkAction<Promise<boolean>, IStoreState, null, AppsAction> {
  return async dispatch => {
    dispatch(requestUpdateInstalledPackage());
    try {
      if (values && schema) {
        const validation = validate(values, schema);
        if (!validation.valid) {
          const errorText =
            validation.errors &&
            validation.errors.map(e => `  - ${e.instancePath}: ${e.message}`).join("\n");
          throw new UnprocessableEntity(
            `The given values don't match the required format. The following errors were found:\n${errorText}`,
          );
        }
      }
      if (availablePackageDetail?.version?.pkgVersion) {
        await App.UpdateInstalledPackage(
          installedPackageRef,
          { version: availablePackageDetail.version.pkgVersion } as VersionReference,
          values,
        );
        dispatch(receiveUpdateInstalledPackage());
        return true;
      } else {
        dispatch(
          errorApp(
            new UpgradeError("This package does not contain enough information to be installed"),
          ),
        );
        return false;
      }
    } catch (e: any) {
      dispatch(errorApp(new UpgradeError(e.message)));
      return false;
    }
  };
}

export function rollbackInstalledPackage(
  installedPackageRef: InstalledPackageReference,
  revision: number,
): ThunkAction<Promise<boolean>, IStoreState, null, AppsAction> {
  return async dispatch => {
    // rollbackInstalledPackage is currently only available for Helm packages
    if (installedPackageRef?.plugin?.name === PluginNames.PACKAGES_HELM) {
      dispatch(requestRollbackInstalledPackage());
      try {
        await App.RollbackInstalledPackage(installedPackageRef, revision);
        dispatch(receiveRollbackInstalledPackage());
        dispatch(getApp(installedPackageRef));
        return true;
      } catch (e: any) {
        dispatch(errorApp(new RollbackError(e.message)));
        return false;
      }
    } else {
      dispatch(
        errorApp(
          new RollbackError(
            "This package cannot be rolled back; this operation is only available for Helm packages",
          ),
        ),
      );
      return false;
    }
  };
}
