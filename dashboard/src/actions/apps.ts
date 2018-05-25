import { Dispatch } from "redux";
import { createAction, getReturnOfExpression } from "typesafe-actions";

import { App } from "../shared/App";
import { HelmRelease } from "../shared/HelmRelease";
import { IApp, IChartVersion, IStoreState } from "../shared/types";

export const requestApps = createAction("REQUEST_APPS");
export const receiveApps = createAction("RECEIVE_APPS", (apps: IApp[]) => {
  return {
    apps,
    type: "RECEIVE_APPS",
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
export const selectApp = createAction("SELECT_APP", (app: IApp) => {
  return {
    app,
    type: "SELECT_APP",
  };
});

const allActions = [requestApps, receiveApps, errorApps, errorDeleteApp, selectApp].map(
  getReturnOfExpression,
);
export type AppsAction = typeof allActions[number];

export function getApp(hrName: string, releaseName: string, namespace: string) {
  return async (dispatch: Dispatch<IStoreState>): Promise<void> => {
    dispatch(requestApps());
    try {
      if (!hrName) {
        hrName = await HelmRelease.getHelmRelease(releaseName, namespace);
      }
      const app = await HelmRelease.getDetails(hrName, releaseName, namespace);
      dispatch(selectApp(app));
    } catch (e) {
      dispatch(errorApps(e));
    }
  };
}

export function deleteApp(releaseName: string, namespace: string) {
  return async (dispatch: Dispatch<IStoreState>): Promise<boolean> => {
    try {
      const hrName = await HelmRelease.getHelmRelease(releaseName, namespace);
      await HelmRelease.delete(hrName, namespace);
      await App.waitForDeletion(releaseName);
      return true;
    } catch (e) {
      dispatch(errorDeleteApp(e));
      return false;
    }
  };
}

export function fetchApps(ns?: string) {
  return async (dispatch: Dispatch<IStoreState>): Promise<void> => {
    if (ns && ns === "_all") {
      ns = undefined;
    }
    dispatch(requestApps());
    try {
      const apps = await HelmRelease.getAllWithDetails(ns);
      dispatch(receiveApps(apps));
    } catch (e) {
      dispatch(errorApps(e));
    }
  };
}

export function deployChart(
  name: string,
  chartVersion: IChartVersion,
  releaseName: string,
  namespace: string,
  values?: string,
  resourceVersion?: string,
) {
  return async (dispatch: Dispatch<IStoreState>): Promise<boolean> => {
    try {
      if (resourceVersion) {
        await HelmRelease.upgrade(name, releaseName, namespace, chartVersion, values);
      } else {
        await HelmRelease.create(name, releaseName, namespace, chartVersion, values);
      }
      return true;
    } catch (e) {
      dispatch(errorApps(e));
      return false;
    }
  };
}
