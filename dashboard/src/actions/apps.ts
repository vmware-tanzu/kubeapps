import { Dispatch } from "redux";
import { createAction, getReturnOfExpression } from "typesafe-actions";

import { App } from "../shared/App";
import { hapi } from "../shared/hapi/release";
import { IAppOverview, IChartVersion, IStoreState } from "../shared/types";

export const requestApps = createAction("REQUEST_APPS");
export const receiveApps = createAction("RECEIVE_APPS", (apps: hapi.release.Release[]) => {
  return {
    apps,
    type: "RECEIVE_APPS",
  };
});
export const listApps = createAction("REQUEST_APP_LIST");
export const receiveAppList = createAction("RECEIVE_APP_LIST", (apps: IAppOverview[]) => {
  return {
    apps,
    type: "RECEIVE_APP_LIST",
  };
});
export const errorApps = createAction("ERROR_APPS", (err: Error) => ({
  err,
  type: "ERROR_APPS",
}));
export const errorDeleteApp = createAction("ERROR_DELETE_APP", (err: Error) => ({
  err,
  type: "ERROR_DELETE_APP",
}));
export const selectApp = createAction("SELECT_APP", (app: hapi.release.Release) => {
  return {
    app,
    type: "SELECT_APP",
  };
});
export const toggleListAllAction = createAction("REQUEST_TOGGLE_LIST_ALL");

const allActions = [
  listApps,
  requestApps,
  receiveApps,
  receiveAppList,
  toggleListAllAction,
  errorApps,
  errorDeleteApp,
  selectApp,
].map(getReturnOfExpression);
export type AppsAction = typeof allActions[number];

export function getApp(releaseName: string, namespace: string) {
  return async (dispatch: Dispatch<IStoreState>): Promise<void> => {
    dispatch(requestApps());
    try {
      const app = await App.getRelease(namespace, releaseName);
      dispatch(selectApp(app));
    } catch (e) {
      dispatch(errorApps(e));
    }
  };
}

export function deleteApp(releaseName: string, namespace: string) {
  return async (dispatch: Dispatch<IStoreState>): Promise<boolean> => {
    try {
      await App.delete(releaseName, namespace);
      return true;
    } catch (e) {
      dispatch(errorDeleteApp(e));
      return false;
    }
  };
}

export function fetchApps(ns?: string, all?: boolean) {
  return async (dispatch: Dispatch<IStoreState>): Promise<void> => {
    if (ns && ns === "_all") {
      ns = undefined;
    }
    dispatch(listApps());
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
) {
  return async (dispatch: Dispatch<IStoreState>, getState: () => IStoreState): Promise<boolean> => {
    try {
      const { config: { namespace: kubeappsNamespace } } = getState();
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
) {
  return async (dispatch: Dispatch<IStoreState>, getState: () => IStoreState): Promise<boolean> => {
    try {
      const { config: { namespace: kubeappsNamespace } } = getState();
      await App.upgrade(releaseName, namespace, kubeappsNamespace, chartVersion, values);
      return true;
    } catch (e) {
      dispatch(errorApps(e));
      return false;
    }
  };
}

export function toggleListAll() {
  return async (dispatch: Dispatch<IStoreState>): Promise<void> => {
    dispatch(toggleListAllAction());
  };
}
