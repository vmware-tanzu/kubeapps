import { JSONSchemaType } from "ajv";
import {
  AvailablePackageDetail,
  InstalledPackageDetail,
  InstalledPackageSummary,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { ThunkAction } from "redux-thunk";
import {
  CreateError,
  DeleteError,
  FetchError,
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

export const receiveApps = createAction("RECEIVE_APPS", resolve => {
  return (apps: InstalledPackageDetail[]) => resolve(apps);
});

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
  return (app: InstalledPackageDetail) => resolve(app);
});

const allActions = [
  listApps,
  requestApps,
  receiveApps,
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
  cluster: string,
  namespace: string,
  releaseName: string,
): ThunkAction<Promise<InstalledPackageDetail | undefined>, IStoreState, null, AppsAction> {
  return async dispatch => {
    dispatch(requestApps());
    try {
      const { installedPackageDetail } = await App.GetInstalledPackageDetail(
        cluster,
        namespace,
        releaseName,
      );
      if (installedPackageDetail) {
        dispatch(selectApp(installedPackageDetail));
      }
      return installedPackageDetail;
    } catch (e) {
      dispatch(errorApp(new FetchError(e.message)));
      return;
    }
  };
}

export function deleteApp(
  cluster: string,
  namespace: string,
  releaseName: string,
  purge: boolean,
): ThunkAction<Promise<boolean>, IStoreState, null, AppsAction> {
  return async dispatch => {
    dispatch(requestDeleteApp());
    try {
      await App.delete(cluster, namespace, releaseName, purge);
      dispatch(receiveDeleteApp());
      return true;
    } catch (e) {
      dispatch(errorApp(new DeleteError(e.message)));
      return false;
    }
  };
}

// fetchApps returns a list of apps for other actions to compose on top of it
export function fetchApps(
  cluster: string,
  ns?: string,
): ThunkAction<Promise<InstalledPackageSummary[]>, IStoreState, null, AppsAction> {
  return async dispatch => {
    dispatch(listApps());
    try {
      const { installedPackageSummaries } = await App.GetInstalledPackageSummaries(cluster, ns);
      dispatch(receiveAppList(installedPackageSummaries));
      return installedPackageSummaries;
    } catch (e) {
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
  return async (dispatch, getState) => {
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
    } catch (e) {
      dispatch(errorApp(new CreateError(e.message)));
      return false;
    }
  };
}

export function upgradeApp(
  cluster: string,
  namespace: string,
  chartVersion: AvailablePackageDetail,
  chartNamespace: string,
  releaseName: string,
  values?: string,
  schema?: JSONSchemaType<any>,
): ThunkAction<Promise<boolean>, IStoreState, null, AppsAction> {
  return async (dispatch, getState) => {
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
      await App.upgrade(cluster, namespace, releaseName, chartNamespace, chartVersion, values);
      dispatch(receiveUpgradeApp());
      return true;
    } catch (e) {
      dispatch(errorApp(new UpgradeError(e.message)));
      return false;
    }
  };
}

export function rollbackApp(
  cluster: string,
  namespace: string,
  releaseName: string,
  revision: number,
): ThunkAction<Promise<boolean>, IStoreState, null, AppsAction> {
  return async (dispatch, getState) => {
    dispatch(requestRollbackApp());
    try {
      await App.rollback(cluster, namespace, releaseName, revision);
      dispatch(receiveRollbackApp());
      // dispatch(getAppWithUpdateInfo(cluster, namespace, releaseName));
      return true;
    } catch (e) {
      dispatch(errorApp(new RollbackError(e.message)));
      return false;
    }
  };
}
