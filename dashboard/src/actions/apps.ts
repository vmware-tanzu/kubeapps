import { JSONSchemaType } from "ajv";
import {
  AvailablePackageDetail,
  InstalledPackageDetail,
  InstalledPackageReference,
  InstalledPackageSummary,
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
import { ActionType, deprecated } from "typesafe-actions";
import { App } from "../shared/App";
import { validate } from "../shared/schema";

const { createAction } = deprecated;

export const requestApps = createAction("REQUEST_APPS");

export const listApps = createAction("REQUEST_APP_LIST");

export const receiveAppList = createAction("RECEIVE_APP_LIST", resolve => {
  return (apps: InstalledPackageSummary[]) => resolve(apps);
});

export const requestDeleteApp = createAction("REQUEST_DELETE_APP");

export const receiveDeleteApp = createAction("RECEIVE_DELETE_APP_CONFIRMATION");

export const requestDeployApp = createAction("REQUEST_DEPLOY_APP");

export const receiveDeployApp = createAction("RECEIVE_DEPLOY_APP_CONFIRMATION");

export const requestUpgradeApp = createAction("REQUEST_UPGRADE_APP");

export const receiveUpgradeApp = createAction("RECEIVE_UPGRADE_APP_CONFIRMATION");

export const requestRollbackApp = createAction("REQUEST_ROLLBACK_APP");

export const receiveRollbackApp = createAction("RECEIVE_ROLLBACK_APP_CONFIRMATION");

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
  requestDeleteApp,
  receiveDeleteApp,
  requestDeployApp,
  receiveDeployApp,
  requestUpgradeApp,
  receiveUpgradeApp,
  requestRollbackApp,
  receiveRollbackApp,
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
      // TODO(agamez): remove it once we return the generated resources as part of the InstalledPackageDetail.
      const legacyResponse = await App.getRelease(installedPackageRef);
      // Get the details of an installed package
      const { installedPackageDetail } = await App.GetInstalledPackageDetail(installedPackageRef);
      // For local packages with no references to any available packages (eg.a local chart for development)
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

export function deleteApp(
  installedPackageRef: InstalledPackageReference,
  purge: boolean,
): ThunkAction<Promise<boolean>, IStoreState, null, AppsAction> {
  return async dispatch => {
    dispatch(requestDeleteApp());
    try {
      await App.delete(installedPackageRef, purge);
      dispatch(receiveDeleteApp());
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
    let installedPackageSummaries;
    try {
      const res = await App.GetInstalledPackageSummaries(cluster, namespace);
      installedPackageSummaries = res?.installedPackageSummaries;
      dispatch(receiveAppList(installedPackageSummaries));
      return installedPackageSummaries;
    } catch (e: any) {
      dispatch(errorApp(new FetchError(e.message)));
      return [];
    }
  };
}

export function deployChart(
  targetCluster: string,
  targetNamespace: string,
  availablePackageDetail: AvailablePackageDetail,
  releaseName: string,
  values?: string,
  schema?: JSONSchemaType<any>,
): ThunkAction<Promise<boolean>, IStoreState, null, AppsAction> {
  return async dispatch => {
    dispatch(requestDeployApp());
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

      await App.create(targetCluster, targetNamespace, releaseName, availablePackageDetail, values);
      dispatch(receiveDeployApp());

      return true;
    } catch (e: any) {
      dispatch(errorApp(new CreateError(e.message)));
      return false;
    }
  };
}

export function upgradeApp(
  installedPackageRef: InstalledPackageReference,
  availablePackageDetail: AvailablePackageDetail,
  chartNamespace: string,
  values?: string,
  schema?: JSONSchemaType<any>,
): ThunkAction<Promise<boolean>, IStoreState, null, AppsAction> {
  return async dispatch => {
    dispatch(requestUpgradeApp());
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
      await App.upgrade(installedPackageRef, chartNamespace, availablePackageDetail, values);
      dispatch(receiveUpgradeApp());
      return true;
    } catch (e: any) {
      dispatch(errorApp(new UpgradeError(e.message)));
      return false;
    }
  };
}

export function rollbackApp(
  installedPackageRef: InstalledPackageReference,
  revision: number,
): ThunkAction<Promise<boolean>, IStoreState, null, AppsAction> {
  return async dispatch => {
    dispatch(requestRollbackApp());
    try {
      await App.rollback(installedPackageRef, revision);
      dispatch(receiveRollbackApp());
      dispatch(getApp(installedPackageRef));
      return true;
    } catch (e: any) {
      dispatch(errorApp(new RollbackError(e.message)));
      return false;
    }
  };
}
