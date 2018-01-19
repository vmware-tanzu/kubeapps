import { Dispatch } from "redux";
import { createAction, getReturnOfExpression } from "typesafe-actions";

import { IChart, IChartVersion, IStoreState } from "../shared/types";
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
export const selectChart = createAction("SELECT_CHART", (chart: IChart) => ({
  chart,
  type: "SELECT_CHART",
}));
export const selectReadme = createAction("SELECT_README", (readme: string) => ({
  readme,
  type: "SELECT_README",
}));

const allActions = [
  requestCharts,
  receiveCharts,
  receiveChartVersions,
  selectChart,
  selectReadme,
].map(getReturnOfExpression);
export type ChartsAction = typeof allActions[number];

export function fetchCharts(repo: string) {
  return (dispatch: Dispatch<IStoreState>): Promise<{}> => {
    dispatch(requestCharts());
    return fetch(url.api.charts.list(repo))
      .then(response => response.json())
      .then(json => dispatch(receiveCharts(json.data)));
  };
}

export function getChart(id: string) {
  return (dispatch: Dispatch<IStoreState>): Promise<{}> => {
    dispatch(requestCharts());
    return fetch(url.api.charts.get(id))
      .then(response => response.json())
      .then(json => {
        const c: IChart = json.data;
        dispatch(listChartVersions(c.id));
        dispatch(getChartReadme(c.id, c.relationships.latestChartVersion.data.version));
        return dispatch(selectChart(json.data));
      });
  };
}

export function listChartVersions(id: string) {
  return (dispatch: Dispatch<IStoreState>): Promise<{}> => {
    return fetch(url.api.charts.listVersions(id))
      .then(response => response.json())
      .then(json => dispatch(receiveChartVersions(json.data)));
  };
}

export function getChartReadme(id: string, version: string) {
  return (dispatch: Dispatch<IStoreState>): Promise<{}> => {
    return fetch(url.api.charts.getReadme(id, version))
      .then(response => response.text())
      .then(text => dispatch(selectReadme(text)));
  };
}

export function deployChart(chart: IChart, releaseName: string, namespace: string) {
  return (dispatch: Dispatch<IStoreState>): Promise<{}> => {
    return fetch(url.api.helmreleases.create(namespace), {
      headers: { "Content-Type": "application/json" },
      method: "POST",

      body: JSON.stringify({
        apiVersion: "helm.bitnami.com/v1",
        kind: "HelmRelease",
        metadata: {
          name: releaseName,
        },
        spec: {
          chartName: chart.attributes.name,
          repoUrl: chart.attributes.repo.url,
          version: chart.relationships.latestChartVersion.data.version,
        },
      }),
    })
      .then(response => response.json())
      .then(json => {
        if (json.status === "Failure") {
          throw new Error(json.message);
        }
        return json;
      });
  };
}
