import { Dispatch } from "redux";
import { createAction, getReturnOfExpression } from "typesafe-actions";

import { App } from "../shared/App";
import Chart from "../shared/Chart";
import { HelmRelease } from "../shared/HelmRelease";
import { IApp, IChartVersion, IStoreState, MissingChart } from "../shared/types";

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

export function getApp(helmCRDReleaseName: string, tillerReleaseName: string, namespace: string) {
  return async (dispatch: Dispatch<IStoreState>): Promise<void> => {
    dispatch(requestApps());
    try {
      if (!helmCRDReleaseName) {
        helmCRDReleaseName = await HelmRelease.getHelmRelease(tillerReleaseName, namespace);
      }
      const app = await HelmRelease.getDetails(helmCRDReleaseName, tillerReleaseName, namespace);
      dispatch(selectApp(app));
    } catch (e) {
      dispatch(errorApps(e));
    }
  };
}

export function deleteApp(tillerReleaseName: string, namespace: string) {
  return async (dispatch: Dispatch<IStoreState>): Promise<boolean> => {
    try {
      const helmCRDReleaseName = await HelmRelease.getHelmRelease(tillerReleaseName, namespace);
      await HelmRelease.delete(helmCRDReleaseName, namespace);
      await App.waitForDeletion(tillerReleaseName);
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
  tillerReleaseName: string,
  namespace: string,
  values?: string,
  resourceVersion?: string,
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
      if (resourceVersion) {
        await HelmRelease.upgrade(name, tillerReleaseName, namespace, chartVersion, values);
      } else {
        await HelmRelease.create(name, tillerReleaseName, namespace, chartVersion, values);
      }
      return true;
    } catch (e) {
      dispatch(errorApps(e));
      return false;
    }
  };
}
