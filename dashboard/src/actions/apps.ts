import { Action, Dispatch } from "redux";
import { ThunkAction } from "redux-thunk";
import { ActionType, createActionDeprecated } from "typesafe-actions";
import { App } from "../shared/App";
import { hapi } from "../shared/hapi/release";
import { definedNamespaces } from "../shared/Namespace";
import { IAppOverview, IChartVersion, IStoreState, UnprocessableEntity } from "../shared/types";

export const requestApps = createActionDeprecated("REQUEST_APPS");
export const receiveApps = createActionDeprecated(
  "RECEIVE_APPS",
  (apps: hapi.release.Release[]) => {
    return {
      apps,
      type: "RECEIVE_APPS",
    };
  },
);
export const listApps = createActionDeprecated("REQUEST_APP_LIST", (listingAll: boolean) => {
  return {
    listingAll,
    type: "REQUEST_APP_LIST",
  };
});
export const receiveAppList = createActionDeprecated("RECEIVE_APP_LIST", (apps: IAppOverview[]) => {
  return {
    apps,
    type: "RECEIVE_APP_LIST",
  };
});
export const errorApps = createActionDeprecated("ERROR_APPS", (err: Error) => ({
  err,
  type: "ERROR_APPS",
}));
export const errorDeleteApp = createActionDeprecated("ERROR_DELETE_APP", (err: Error) => ({
  err,
  type: "ERROR_DELETE_APP",
}));
export const selectApp = createActionDeprecated("SELECT_APP", (app: hapi.release.Release) => {
  return {
    app,
    type: "SELECT_APP",
  };
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

export function getApp(releaseName: string, namespace: string) {
  return async (dispatch: Dispatch): Promise<void> => {
    dispatch(requestApps());
    try {
      const app = await App.getRelease(namespace, releaseName);
      dispatch(selectApp(app));
    } catch (e) {
      dispatch(errorApps(e));
    }
  };
}

export function deleteApp(releaseName: string, namespace: string, purge: boolean) {
  return async (dispatch: Dispatch): Promise<boolean> => {
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
): ThunkAction<Promise<void>, IStoreState, null, Action> {
  return async (dispatch: Dispatch): Promise<void> => {
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
) {
  return async (dispatch: Dispatch, getState: () => IStoreState): Promise<boolean> => {
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
) {
  return async (dispatch: Dispatch, getState: () => IStoreState): Promise<boolean> => {
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
