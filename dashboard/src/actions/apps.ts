import { JSONSchemaType } from "ajv";
import {
  AvailablePackageDetail,
  Context,
  InstalledPackageDetail,
  InstalledPackageReference,
  InstalledPackageSummary,
  ReconciliationOptions,
  ResourceRef,
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
import { getPluginsSupportingRollback } from "shared/utils";
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

export const clearErrorApp = createAction("CLEAR_ERROR_APP");

export const selectApp = createAction("SELECT_APP", resolve => {
  return (
    app: InstalledPackageDetail,
    resourceRefs: ResourceRef[],
    details?: AvailablePackageDetail,
  ) => resolve({ app, resourceRefs, details });
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
  clearErrorApp,
  selectApp,
];

export type AppsAction = ActionType<typeof allActions[number]>;

export function getApp(
  installedPackageRef?: InstalledPackageReference,
): ThunkAction<Promise<void>, IStoreState, null, AppsAction> {
  return async dispatch => {
    dispatch(requestApps());
    try {
      // Get the details of an installed package
      const { installedPackageDetail } = await App.GetInstalledPackageDetail(installedPackageRef);
      const { resourceRefs } = await App.GetInstalledPackageResourceRefs(installedPackageRef);

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
      dispatch(selectApp(installedPackageDetail!, resourceRefs, availablePackageDetail));
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

      dispatch(receiveAppList(installedPackageSummaries));
      return installedPackageSummaries;
    } catch (e: any) {
      dispatch(
        errorApp(
          e instanceof Error
            ? new FetchError("Unable to list apps", [e])
            : new FetchError("Unable to list apps: " + e.message),
        ),
      );
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
  reconciliationOptions?: ReconciliationOptions,
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
          reconciliationOptions as ReconciliationOptions,
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
    if (
      installedPackageRef?.plugin?.name &&
      getPluginsSupportingRollback().includes(installedPackageRef.plugin.name)
    ) {
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
