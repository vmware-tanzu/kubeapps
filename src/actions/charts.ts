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
export const selectChartVersion = createAction(
  "SELECT_CHART_VERSION",
  (chartVersion: IChartVersion) => ({
    chartVersion,
    type: "SELECT_CHART_VERSION",
  }),
);
export const selectReadme = createAction("SELECT_README", (readme: string) => ({
  readme,
  type: "SELECT_README",
}));
export const selectValues = createAction("SELECT_VALUES", (values: string) => ({
  type: "SELECT_VALUES",
  values,
}));

const allActions = [
  requestCharts,
  receiveCharts,
  receiveChartVersions,
  selectChartVersion,
  selectReadme,
  selectValues,
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

export function fetchChartVersions(id: string) {
  return (dispatch: Dispatch<IStoreState>): Promise<{}> => {
    dispatch(requestCharts());
    return fetch(url.api.charts.listVersions(id))
      .then(response => response.json())
      .then(json => dispatch(receiveChartVersions(json.data)));
  };
}

export function fetchChartVersionsAndSelectVersion(id: string, version?: string) {
  return (dispatch: Dispatch<IStoreState>): Promise<{}> => {
    return dispatch(fetchChartVersions(id)).then((action: any) => {
      const versions: IChartVersion[] = action.versions;
      let cv: IChartVersion = versions[0];
      if (version) {
        const found = versions.find(v => v.attributes.version === version);
        if (!found) {
          throw new Error("could not find chart version");
        }
        cv = found;
      }
      dispatch(getChartReadme(id, cv.attributes.version));
      dispatch(getChartValues(id, cv.attributes.version));
      return dispatch(selectChartVersion(cv));
    });
  };
}

export function selectChartVersionAndGetFiles(cv: IChartVersion) {
  return (dispatch: Dispatch<IStoreState>): Promise<{}> => {
    const id = `${cv.relationships.chart.data.repo.name}/${cv.relationships.chart.data.name}`;
    dispatch(selectChartVersion(cv));
    dispatch(getChartValues(id, cv.attributes.version));
    return dispatch(getChartReadme(id, cv.attributes.version));
  };
}

export function getChartReadme(id: string, version: string) {
  return (dispatch: Dispatch<IStoreState>): Promise<{}> => {
    return fetch(url.api.charts.getReadme(id, version))
      .then(response => response.text())
      .then(text => dispatch(selectReadme(text)));
  };
}

export function getChartValues(id: string, version: string) {
  return (dispatch: Dispatch<IStoreState>): Promise<{}> => {
    return fetch(url.api.charts.getValues(id, version))
      .then(response => response.text())
      .then(text => dispatch(selectValues(text)));
  };
}

export function deployChart(
  chartVersion: IChartVersion,
  releaseName: string,
  namespace: string,
  values: string,
) {
  return (dispatch: Dispatch<IStoreState>): Promise<{}> => {
    const chartAttrs = chartVersion.relationships.chart.data;
    return fetch(url.api.helmreleases.create(namespace), {
      headers: { "Content-Type": "application/json" },
      method: "POST",

      body: JSON.stringify({
        apiVersion: "helm.bitnami.com/v1",
        kind: "HelmRelease",
        metadata: {
          annotations: {
            "apprepositories.kubeapps.com/repo-name": chartAttrs.repo.name,
          },
          name: releaseName,
        },
        spec: {
          chartName: chartAttrs.name,
          repoUrl: chartAttrs.repo.url,
          values,
          version: chartVersion.attributes.version,
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
