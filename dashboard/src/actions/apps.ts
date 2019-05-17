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

export const requestAppUpdateInfo = createAction("REQUEST_APP_UPDATE_INFO");

export const receiveAppUpdateInfo = createAction("RECEIVE_APP_UPDATE_INFO", resolve => {
  return (payload: { releaseName: string; updateInfo: IChartUpdateInfo }) => resolve(payload);
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
  requestAppUpdateInfo,
  receiveAppUpdateInfo,
  errorApps,
  errorDeleteApp,
  selectApp,
];

export type AppsAction = ActionType<typeof allActions[number]>;

export function getApp(
  releaseName: string,
  namespace: string,
): ThunkAction<Promise<hapi.release.Release | undefined>, IStoreState, null, AppsAction> {
  return async dispatch => {
    dispatch(requestApps());
    try {
      const app = await App.getRelease(namespace, releaseName);
      dispatch(selectApp(app));
      return app;
    } catch (e) {
      dispatch(errorApps(e));
      return;
    }
  };
}

function getAppUpdateInfo(
  releaseName: string,
  chartName: string,
  currentVersion: string,
  appVersion: string,
): ThunkAction<Promise<void>, IStoreState, null, AppsAction> {
  return async dispatch => {
    dispatch(requestAppUpdateInfo());
    try {
      const chartsInfo = await Chart.listWithFilters(chartName, currentVersion, appVersion);
      let updateInfo: IChartUpdateInfo = {
        upToDate: true,
        repository: { name: "", url: "" },
        latestVersion: "",
      };
      if (chartsInfo.length > 0) {
        const sortedCharts = chartsInfo.sort((a, b) =>
          semver.compare(
            a.relationships.latestChartVersion.data.version,
            b.relationships.latestChartVersion.data.version,
          ),
        );
        const latestVersion = sortedCharts[0].relationships.latestChartVersion.data.version;
        // Initialize updateInfo with the latest chart found
        updateInfo = {
          upToDate: semver.gte(currentVersion, latestVersion),
          latestVersion,
          repository: sortedCharts[0].attributes.repo,
        };
      }
      dispatch(receiveAppUpdateInfo({ releaseName, updateInfo }));
    } catch (e) {
      if (e.message.includes("Invalid Version")) {
        // The version is not semver compatible so we cannot look for new versions
        const updateInfo: IChartUpdateInfo = {
          upToDate: true,
          repository: { name: "", url: "" },
          latestVersion: "",
        };
        dispatch(receiveAppUpdateInfo({ releaseName, updateInfo }));
        return;
      }
      dispatch(errorApps(e));
    }
  };
}

export function getAppWithUpdateInfo(
  releaseName: string,
  namespace: string,
): ThunkAction<Promise<void>, IStoreState, null, AppsAction> {
  return async dispatch => {
    dispatch(requestApps());
    try {
      const app = await dispatch(getApp(releaseName, namespace));
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
            app.name,
            app.chart.metadata.name,
            app.chart.metadata.version,
            app.chart.metadata.appVersion,
          ),
        );
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

// fetchApps returns a list of apps for other actions to compose on top of it
export function fetchApps(
  ns?: string,
  all: boolean = false,
): ThunkAction<Promise<IAppOverview[]>, IStoreState, null, AppsAction> {
  return async dispatch => {
    if (ns && ns === definedNamespaces.all) {
      ns = undefined;
    }
    dispatch(listApps(all));
    try {
      const apps = await App.listApps(ns, all);
      dispatch(receiveAppList(apps));
      return apps;
    } catch (e) {
      dispatch(errorApps(e));
      return [];
    }
  };
}

export function fetchAppsWithUpdateInfo(
  ns?: string,
  all: boolean = false,
): ThunkAction<Promise<void>, IStoreState, null, AppsAction> {
  return async dispatch => {
    const apps = await dispatch(fetchApps(ns, all));
    apps.forEach(app =>
      dispatch(
        getAppUpdateInfo(
          app.releaseName,
          app.chartMetadata.name,
          app.chartMetadata.version,
          app.chartMetadata.appVersion,
        ),
      ),
    );
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
