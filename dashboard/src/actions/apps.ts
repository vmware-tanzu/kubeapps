import { ThunkAction } from "redux-thunk";
import { ActionType, createAction } from "typesafe-actions";
import { App } from "../shared/App";
import { hapi } from "../shared/hapi/release";
import { definedNamespaces } from "../shared/Namespace";
import { IAppOverview, IChartVersion, IStoreState, UnprocessableEntity } from "../shared/types";

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

export const errorApps = createAction("ERROR_APPS", resolve => {
  return (err: Error) => resolve(err);
});

export const errorDeleteApp = createAction("ERROR_DELETE_APP", resolve => {
  return (err: Error) => resolve(err);
});

export const selectApp = createAction("SELECT_APP", resolve => {
  return (app: hapi.release.Release) => resolve(app);
});

const allActions = [
  listApps,
  requestApps,
  receiveApps,
  receiveAppList,
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
): ThunkAction<Promise<void>, IStoreState, null, AppsAction> {
  return async dispatch => {
    if (ns && ns === definedNamespaces.all) {
      ns = undefined;
    }
    dispatch(listApps(all));
    try {
      const apps = await App.listApps(ns, all);
      dispatch(receiveAppList(apps));
    } catch (e) {
      dispatch(errorApps(e));
    }
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
