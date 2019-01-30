import { ThunkAction } from "redux-thunk";
import * as semver from "semver";
import { ActionType, createAction } from "typesafe-actions";
import { App } from "../shared/App";
import Chart from "../shared/Chart";
import { hapi } from "../shared/hapi/release";
import { definedNamespaces } from "../shared/Namespace";
import {
  IAppOverview,
  IChartUpdateInfo,
  IChartVersion,
  IRelease,
  IStoreState,
  UnprocessableEntity,
} from "../shared/types";

export const requestApps = createAction("REQUEST_APPS");

export const receiveApps = createAction("RECEIVE_APPS", resolve => {
  return (apps: hapi.release.Release[]) => resolve(apps);
});

export const listApps = createAction("REQUEST_APP_LIST", resolve => {
  return (listingAll: boolean) => resolve(listingAll);
});

export const receiveAppList = createAction("RECEIVE_APP_LIST", resolve => {
  return (apps: IAppOverview[]) => resolve(apps);
});

export const updateAppListItem = createAction("UPDATE_APP_LIST_ITEM", resolve => {
  return (app: IAppOverview) => resolve(app);
});

export const errorApps = createAction("ERROR_APPS", resolve => {
  return (err: Error) => resolve(err);
});

export const errorDeleteApp = createAction("ERROR_DELETE_APP", resolve => {
  return (err: Error) => resolve(err);
});

export const selectApp = createAction("SELECT_APP", resolve => {
  return (app: IRelease) => resolve(app);
});

const allActions = [
  listApps,
  requestApps,
  receiveApps,
  receiveAppList,
  updateAppListItem,
  errorApps,
  errorDeleteApp,
  selectApp,
];

export type AppsAction = ActionType<typeof allActions[number]>;

export function getApp(
  releaseName: string,
  namespace: string,
): ThunkAction<Promise<void>, IStoreState, null, AppsAction> {
  return async dispatch => {
    dispatch(requestApps());
    try {
      const app = await App.getRelease(namespace, releaseName);
      dispatch(selectApp(app));
    } catch (e) {
      dispatch(errorApps(e));
    }
  };
}

async function getAppUpdates(name: string, currentVersion: string, appVersion: string) {
  const chartsInfo = await Chart.listWithFilters(name, currentVersion, appVersion);
  let updateInfo: IChartUpdateInfo = {
    repository: { name: "", url: "" },
    latestVersion: "",
  };
  chartsInfo.forEach(c => {
    const chartLatestVersion = c.relationships.latestChartVersion.data.version;
    if (semver.gt(chartLatestVersion, currentVersion)) {
      if (updateInfo.latestVersion && semver.gt(updateInfo.latestVersion, chartLatestVersion)) {
        // The current update is newer than the chart version, do nothing
      } else {
        updateInfo = {
          latestVersion: chartLatestVersion,
          repository: c.attributes.repo,
        };
      }
    }
  });
  return updateInfo;
}

export function getAppWithUpdateInfo(
  releaseName: string,
  namespace: string,
): ThunkAction<Promise<void>, IStoreState, null, AppsAction> {
  return async dispatch => {
    dispatch(requestApps());
    try {
      const app = await App.getRelease(namespace, releaseName);
      dispatch(selectApp(app));
      if (
        app.chart &&
        app.chart.metadata &&
        app.chart.metadata.name &&
        app.chart.metadata.version &&
        app.chart.metadata.appVersion
      ) {
        const name = app.chart.metadata.name;
        const currentVersion = app.chart.metadata.version;
        const appVersion = app.chart.metadata.appVersion;
        const updateInfo = await getAppUpdates(name, currentVersion, appVersion);
        const appWithUpdateInfo = Object.assign({ updateInfo }, app);
        dispatch(selectApp(appWithUpdateInfo));
      }
    } catch (e) {
      dispatch(errorApps(e));
    }
  };
}

export function deleteApp(
  releaseName: string,
  namespace: string,
  purge: boolean,
): ThunkAction<Promise<boolean>, IStoreState, null, AppsAction> {
  return async dispatch => {
    try {
      await App.delete(releaseName, namespace, purge);
      return true;
    } catch (e) {
      dispatch(errorDeleteApp(e));
      return false;
    }
  };
}

export function fetchApps(
  ns?: string,
  all: boolean = false,
): ThunkAction<Promise<AppsAction>, IStoreState, null, AppsAction> {
  return async dispatch => {
    if (ns && ns === definedNamespaces.all) {
      ns = undefined;
    }
    dispatch(listApps(all));
    try {
      const apps = await App.listApps(ns, all);
      return dispatch(receiveAppList(apps));
    } catch (e) {
      return dispatch(errorApps(e));
    }
  };
}

export function fetchAppsWithUpdatesInfo(
  ns?: string,
  all: boolean = false,
): ThunkAction<Promise<Array<Promise<void>>>, IStoreState, null, AppsAction> {
  return async dispatch => {
    const fetchAction = await dispatch(fetchApps(ns, all));
    let apps: IAppOverview[] = [];
    if (fetchAction.type === "RECEIVE_APP_LIST") {
      apps = fetchAction.payload;
    }
    return apps.map(async app => {
      const name = app.chartMetadata.name;
      const currentVersion = app.chartMetadata.version;
      const appVersion = app.chartMetadata.appVersion;
      const updateInfo = await getAppUpdates(name, currentVersion, appVersion);
      dispatch(updateAppListItem({ ...app, updateInfo }));
    });
  };
}

export function deployChart(
  chartVersion: IChartVersion,
  releaseName: string,
  namespace: string,
  values?: string,
): ThunkAction<Promise<boolean>, IStoreState, null, AppsAction> {
  return async (dispatch, getState) => {
    try {
      // You can not deploy applications unless the namespace is set
      if (namespace === definedNamespaces.all) {
        throw new UnprocessableEntity(
          "Namespace not selected. Please select a namespace using the selector in the top right corner.",
        );
      }

      const {
        config: { namespace: kubeappsNamespace },
      } = getState();
      await App.create(releaseName, namespace, kubeappsNamespace, chartVersion, values);
      return true;
    } catch (e) {
      dispatch(errorApps(e));
      return false;
    }
  };
}

export function upgradeApp(
  chartVersion: IChartVersion,
  releaseName: string,
  namespace: string,
  values?: string,
): ThunkAction<Promise<boolean>, IStoreState, null, AppsAction> {
  return async (dispatch, getState) => {
    try {
      const {
        config: { namespace: kubeappsNamespace },
      } = getState();
      await App.upgrade(releaseName, namespace, kubeappsNamespace, chartVersion, values);
      return true;
    } catch (e) {
      dispatch(errorApps(e));
      return false;
    }
  };
}
