import { Dispatch } from "redux";
import { createAction, getReturnOfExpression } from "typesafe-actions";

import { AppRepository } from "../shared/AppRepository";
import Chart from "../shared/Chart";
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
  receiveCharts,
  receiveChartVersions,
  selectChartVersion,
  resetChartVersion,
  selectReadme,
  errorReadme,
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

export function getChartVersion(id: string, version: string) {
  return (dispatch: Dispatch<IStoreState>): Promise<{}> => {
    dispatch(requestCharts());
    return fetch(url.api.charts.getVersion(id, version))
      .then(response => response.json())
      .then(json => dispatch(selectChartVersion(json.data)));
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
      return dispatch(selectChartVersion(cv));
    });
  };
}

export function selectChartVersionAndGetFiles(cv: IChartVersion) {
  return (dispatch: Dispatch<IStoreState>): Promise<any> => {
    const id = `${cv.relationships.chart.data.repo.name}/${cv.relationships.chart.data.name}`;
    dispatch(selectChartVersion(cv));
    dispatch(getChartValues(id, cv.attributes.version));
    return dispatch(getChartReadme(id, cv.attributes.version));
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

export function deployChart(
  chartVersion: IChartVersion,
  releaseName: string,
  namespace: string,
  values?: string,
  resourceVersion?: string,
) {
  return async (dispatch: Dispatch<IStoreState>): Promise<{}> => {
    const chartAttrs = chartVersion.relationships.chart.data;
    const method = resourceVersion ? "PUT" : "POST";
    const endpoint = resourceVersion
      ? url.api.helmreleases.upgrade(namespace, releaseName)
      : url.api.helmreleases.create(namespace);
    const repo = await AppRepository.get(chartAttrs.repo.name);
    const auth = repo.spec.auth;
    return fetch(endpoint, {
      headers: { "Content-Type": "application/json" },
      method,

      body: JSON.stringify({
        apiVersion: "helm.bitnami.com/v1",
        kind: "HelmRelease",
        metadata: {
          annotations: {
            "apprepositories.kubeapps.com/repo-name": chartAttrs.repo.name,
          },
          name: releaseName,
          resourceVersion,
        },
        spec: {
          auth,
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
