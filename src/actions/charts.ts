import { Dispatch } from "redux";
import { createAction, getReturnOfExpression } from "typesafe-actions";

import { IChart, IStoreState } from "../shared/types";
import * as url from "../shared/url";

export const requestCharts = createAction("REQUEST_CHARTS");
export const receiveCharts = createAction("RECEIVE_CHARTS", (charts: IChart[]) => ({
  charts,
  type: "RECEIVE_CHARTS",
}));
export const selectChart = createAction("SELECT_CHART", (chart: IChart) => ({
  chart,
  type: "SELECT_CHART",
}));

const allActions = [requestCharts, receiveCharts, selectChart].map(getReturnOfExpression);
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
      .then(json => dispatch(selectChart(json.data)));
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
