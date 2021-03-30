import { JSONSchema4 } from "json-schema";
import { ThunkAction } from "redux-thunk";
import * as semver from "semver";
import { ActionType, deprecated } from "typesafe-actions";
import { App } from "../shared/App";
import Chart from "../shared/Chart";
import { hapi } from "../shared/hapi/release";
import { validate } from "../shared/schema";
import {
  CreateError,
  DeleteError,
  FetchError,
  IAppOverview,
  IChartUpdateInfo,
  IChartVersion,
  IRelease,
  IStoreState,
  RollbackError,
  UnprocessableEntity,
  UpgradeError,
} from "../shared/types";

const { createAction } = deprecated;

export const requestApps = createAction("REQUEST_APPS");

export const receiveApps = createAction("RECEIVE_APPS", resolve => {
  return (apps: hapi.release.Release[]) => resolve(apps);
});

export const listApps = createAction("REQUEST_APP_LIST");

export const receiveAppList = createAction("RECEIVE_APP_LIST", resolve => {
  return (apps: IAppOverview[]) => resolve(apps);
});

export const requestAppUpdateInfo = createAction("REQUEST_APP_UPDATE_INFO");

export const receiveAppUpdateInfo = createAction("RECEIVE_APP_UPDATE_INFO", resolve => {
  return (payload: { releaseName: string; updateInfo: IChartUpdateInfo }) => resolve(payload);
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
  return (app: IRelease) => resolve(app);
});

const allActions = [
  listApps,
  requestApps,
  receiveApps,
  receiveAppList,
  requestAppUpdateInfo,
  receiveAppUpdateInfo,
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
): ThunkAction<Promise<hapi.release.Release | undefined>, IStoreState, null, AppsAction> {
  return async dispatch => {
    dispatch(requestApps());
    try {
      const app = await App.getRelease(cluster, namespace, releaseName);
      dispatch(selectApp(app));
      return app;
    } catch (e) {
      dispatch(errorApp(new FetchError(e.message)));
      return;
    }
  };
}

function getAppUpdateInfo(
  cluster: string,
  namespace: string,
  releaseName: string,
  chartName: string,
  currentVersion: string,
  appVersion: string,
): ThunkAction<Promise<void>, IStoreState, null, AppsAction> {
  return async (dispatch, getState) => {
    dispatch(requestAppUpdateInfo());
    try {
      const chartsInfo = await Chart.listWithFilters(
        cluster,
        namespace,
        chartName,
        currentVersion,
        appVersion,
      );
      let updateInfo: IChartUpdateInfo = {
        upToDate: true,
        repository: { name: "", url: "", namespace: "" },
        chartLatestVersion: "",
        appLatestVersion: "",
      };
      if (chartsInfo.length > 0) {
        const sortedCharts = chartsInfo.sort((a, b) =>
          semver.compare(
            a.relationships.latestChartVersion.data.version,
            b.relationships.latestChartVersion.data.version,
          ),
        );
        const chartLatestVersion = sortedCharts[0].relationships.latestChartVersion.data.version;
        const appLatestVersion = sortedCharts[0].relationships.latestChartVersion.data.app_version;
        // Initialize updateInfo with the latest chart found
        updateInfo = {
          upToDate: semver.gte(currentVersion, chartLatestVersion),
          chartLatestVersion,
          appLatestVersion,
          repository: sortedCharts[0].attributes.repo,
        };
      }
      dispatch(receiveAppUpdateInfo({ releaseName, updateInfo }));
    } catch (e) {
      const updateInfo: IChartUpdateInfo = {
        error: e,
        upToDate: false,
        repository: { name: "", url: "", namespace: "" },
        chartLatestVersion: "",
        appLatestVersion: "",
      };
      dispatch(receiveAppUpdateInfo({ releaseName, updateInfo }));
    }
  };
}

export function getAppWithUpdateInfo(
  cluster: string,
  namespace: string,
  releaseName: string,
): ThunkAction<Promise<void>, IStoreState, null, AppsAction> {
  return async dispatch => {
    try {
      const app = await dispatch(getApp(cluster, namespace, releaseName));
      if (
        app &&
        app.chart &&
        app.chart.metadata &&
        app.chart.metadata.name &&
        app.chart.metadata.version &&
        app.chart.metadata.appVersion
      ) {
        dispatch(
          getAppUpdateInfo(
            cluster,
            namespace,
            app.name,
            app.chart.metadata.name,
            app.chart.metadata.version,
            app.chart.metadata.appVersion,
          ),
        );
      }
    } catch (e) {
      dispatch(errorApp(new FetchError(e.message)));
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
): ThunkAction<Promise<IAppOverview[]>, IStoreState, null, AppsAction> {
  return async dispatch => {
    dispatch(listApps());
    try {
      const apps = await App.listApps(cluster, ns);
      dispatch(receiveAppList(apps));
      return apps;
    } catch (e) {
      dispatch(errorApp(new FetchError(e.message)));
      return [];
    }
  };
}

export function fetchAppsWithUpdateInfo(
  cluster: string,
  namespaceOrAll: string,
): ThunkAction<Promise<void>, IStoreState, null, AppsAction> {
  return async dispatch => {
    try {
      const apps = await dispatch(fetchApps(cluster, namespaceOrAll));
      apps?.forEach(app =>
        dispatch(
          getAppUpdateInfo(
            cluster,
            app.namespace,
            app.releaseName,
            app.chartMetadata.name,
            app.chartMetadata.version,
            app.chartMetadata.appVersion,
          ),
        ),
      );
    } catch (e) {
      dispatch(errorApp(new FetchError(e.message)));
    }
  };
}

export function deployChart(
  targetCluster: string,
  targetNamespace: string,
  chartVersion: IChartVersion,
  chartNamespace: string,
  releaseName: string,
  values?: string,
  schema?: JSONSchema4,
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
      await App.create(
        targetCluster,
        targetNamespace,
        releaseName,
        chartNamespace,
        chartVersion,
        values,
      );
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
  chartVersion: IChartVersion,
  chartNamespace: string,
  releaseName: string,
  values?: string,
  schema?: JSONSchema4,
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
      dispatch(getAppWithUpdateInfo(cluster, namespace, releaseName));
      return true;
    } catch (e) {
      dispatch(errorApp(new RollbackError(e.message)));
      return false;
    }
  };
}
