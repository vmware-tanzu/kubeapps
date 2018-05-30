import { Dispatch } from "redux";
import { createAction, getReturnOfExpression } from "typesafe-actions";

import { App } from "../shared/App";
import Chart from "../shared/Chart";
import { HelmRelease } from "../shared/HelmRelease";
import { AppConflict, IApp, IChartVersion, IStoreState, MissingChart } from "../shared/types";

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

export function getApp(releaseName: string, namespace: string) {
  return async (dispatch: Dispatch<IStoreState>): Promise<void> => {
    dispatch(requestApps());
    try {
      const app = await HelmRelease.getDetails(releaseName, namespace);
      dispatch(selectApp(app));
    } catch (e) {
      dispatch(errorApps(e));
    }
  };
}

export function deleteApp(releaseName: string, namespace: string) {
  return async (dispatch: Dispatch<IStoreState>): Promise<boolean> => {
    try {
      await HelmRelease.delete(releaseName, namespace);
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
  chartVersion: IChartVersion,
  releaseName: string,
  namespace: string,
  values?: string,
  resourceVersion?: string,
) {
  return async (dispatch: Dispatch<IStoreState>): Promise<boolean> => {
    try {
      if (resourceVersion) {
        await HelmRelease.upgrade(releaseName, namespace, chartVersion, values);
      } else {
        const releaseExists = await App.exists(releaseName);
        if (releaseExists) {
          dispatch(errorApps(new AppConflict("Already exists")));
          return false;
        }
        await HelmRelease.create(releaseName, namespace, chartVersion, values);
      }
      return true;
    } catch (e) {
      dispatch(errorApps(e));
      return false;
    }
  };
}

export function migrateApp(
  chartVersion: IChartVersion,
  releaseName: string,
  namespace: string,
  values?: string,
) {
  return async (dispatch: Dispatch<IStoreState>): Promise<boolean> => {
    try {
      const chartExists = await Chart.exists(
        chartVersion.relationships.chart.data.name,
        chartVersion.attributes.version,
        chartVersion.relationships.chart.data.repo.name,
      );
      if (!chartExists) {
        dispatch(errorApps(new MissingChart("Not found")));
        return false;
      }
      await HelmRelease.create(releaseName, namespace, chartVersion, values);
      return true;
    } catch (e) {
      dispatch(errorApps(e));
      return false;
    }
  };
}
