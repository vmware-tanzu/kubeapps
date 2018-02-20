import { Dispatch } from "redux";
import { createAction, getReturnOfExpression } from "typesafe-actions";

import { HelmRelease } from "../shared/HelmRelease";
import { IApp, IStoreState } from "../shared/types";

export const requestApps = createAction("REQUEST_APPS");
export const receiveApps = createAction("RECEIVE_APPS", (apps: IApp[]) => {
  return {
    apps,
    type: "RECEIVE_APPS",
  };
});
export const selectApp = createAction("SELECT_APP", (app: IApp) => {
  return {
    app,
    type: "SELECT_APP",
  };
});

const allActions = [requestApps, receiveApps, selectApp].map(getReturnOfExpression);
export type AppsAction = typeof allActions[number];

export function getApp(releaseName: string, namespace: string) {
  return async (dispatch: Dispatch<IStoreState>): Promise<void> => {
    const app = await HelmRelease.getDetails(releaseName, namespace);
    dispatch(selectApp(app));
  };
}

export function deleteApp(releaseName: string, namespace: string) {
  return async (dispatch: Dispatch<IStoreState>): Promise<void> => {
    return await HelmRelease.delete(releaseName, namespace);
  };
}

export function fetchApps() {
  return async (dispatch: Dispatch<IStoreState>): Promise<void> => {
    dispatch(requestApps());
    const apps = await HelmRelease.getAllWithDetails();
    dispatch(receiveApps(apps));
  };
}
