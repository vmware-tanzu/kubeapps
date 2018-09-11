import { Dispatch } from "redux";
import { createAction, getReturnOfExpression } from "typesafe-actions";

import Chart from "../shared/Chart";
import { IChart, IChartVersion, IStoreState, NotFoundError } from "../shared/types";
import * as url from "../shared/url";

export const requestCharts = createAction("REQUEST_CHARTS");
export const receiveCharts = createAction("RECEIVE_CHARTS", (charts: IChart[]) => ({
  charts,
  type: "RECEIVE_CHARTS",
}));
export const receiveChartVersions = createAction(
  "RECEIVE_CHART_VERSIONS",
  (versions: IChartVersion[]) => ({
    type: "RECEIVE_CHART_VERSIONS",
    versions,
  }),
);
export const errorChart = createAction("ERROR_CHART", (err: Error) => ({
  err,
  type: "ERROR_CHART",
}));
export const selectChartVersion = createAction(
  "SELECT_CHART_VERSION",
  (chartVersion: IChartVersion) => ({
    chartVersion,
    type: "SELECT_CHART_VERSION",
  }),
);
export const resetChartVersion = createAction("RESET_CHART_VERSION", () => ({
  type: "RESET_CHART_VERSION",
}));
export const selectReadme = createAction("SELECT_README", (readme: string) => ({
  readme,
  type: "SELECT_README",
}));
export const errorReadme = createAction("ERROR_README", (message: string) => ({
  message,
  type: "ERROR_README",
}));
export const selectValues = createAction("SELECT_VALUES", (values: string) => ({
  type: "SELECT_VALUES",
  values,
}));

const allActions = [
  requestCharts,
  errorChart,
  receiveCharts,
  receiveChartVersions,
  selectChartVersion,
  resetChartVersion,
  selectReadme,
  errorReadme,
  selectValues,
].map(getReturnOfExpression);
export type ChartsAction = typeof allActions[number];

async function httpGet(dispatch: Dispatch<IStoreState>, targetURL: string) {
  try {
    const response = await fetch(targetURL);
    const json = await response.json();
    if (!response.ok) {
      const error = json.data || response.statusText;
      if (response.status === 404) {
        dispatch(errorChart(new NotFoundError(error)));
      } else {
        dispatch(errorChart(new Error(error)));
      }
    } else {
      return json.data;
    }
  } catch (e) {
    dispatch(errorChart(e));
  }
}

export function fetchCharts(repo: string) {
  return async (dispatch: Dispatch<IStoreState>): Promise<{}> => {
    dispatch(requestCharts());
    const response = await httpGet(dispatch, url.api.charts.list(repo));
    if (response) {
      dispatch(receiveCharts(response));
    }
    return response;
  };
}

export function fetchChartVersions(id: string) {
  return async (dispatch: Dispatch<IStoreState>): Promise<{}> => {
    dispatch(requestCharts());
    const response = await httpGet(dispatch, url.api.charts.listVersions(id));
    if (response) {
      dispatch(receiveChartVersions(response));
    }
    return response;
  };
}

export function getChartVersion(id: string, version: string) {
  return async (dispatch: Dispatch<IStoreState>): Promise<{}> => {
    dispatch(requestCharts());
    const response = await httpGet(dispatch, url.api.charts.getVersion(id, version));
    if (response) {
      dispatch(selectChartVersion(response));
    }
    return response;
  };
}

export function fetchChartVersionsAndSelectVersion(id: string, version?: string) {
  return async (dispatch: Dispatch<IStoreState>) => {
    const versions = (await dispatch(fetchChartVersions(id))) as IChartVersion[];
    if (versions) {
      let cv: IChartVersion = versions[0];
      if (version) {
        const found = versions.find(v => v.attributes.version === version);
        if (!found) {
          throw new Error("could not find chart version");
        }
        cv = found;
      }
      dispatch(selectChartVersion(cv));
    }
  };
}

export function getChartReadme(id: string, version: string) {
  return async (dispatch: Dispatch<IStoreState>) => {
    try {
      const readme = await Chart.getReadme(id, version);
      dispatch(selectReadme(readme));
      return readme;
    } catch (e) {
      return dispatch(errorReadme(e.toString()));
    }
  };
}

export function getChartValues(id: string, version: string) {
  return async (dispatch: Dispatch<IStoreState>) => {
    try {
      const values = await Chart.getValues(id, version);
      dispatch(selectValues(values));
      return values;
    } catch (e) {
      dispatch(selectValues(""));
      return "";
    }
  };
}
