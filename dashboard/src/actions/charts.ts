import { JSONSchema4 } from "json-schema";
import { ThunkAction } from "redux-thunk";
import * as semver from "semver";
import { ActionType, createAction } from "typesafe-actions";

import Chart from "../shared/Chart";
import {
  FetchError,
  IChartCategory,
  IChartVersion,
  IReceiveChartsActionPayload,
  IStoreState,
  NotFoundError,
} from "../shared/types";

export const requestCharts = createAction("REQUEST_CHARTS", resolve => {
  return (page?: number) => resolve(page);
});

export const requestChart = createAction("REQUEST_CHART");

export const receiveCharts = createAction("RECEIVE_CHARTS", resolve => {
  return (payload: IReceiveChartsActionPayload) => resolve(payload);
});

export const requestChartsCategories = createAction("REQUEST_CHARTS_CATEGORIES");

export const receiveChartCategories = createAction("RECEIVE_CHART_CATEGORIES", resolve => {
  return (categories: IChartCategory[]) => resolve(categories);
});

export const receiveChartVersions = createAction("RECEIVE_CHART_VERSIONS", resolve => {
  return (versions: IChartVersion[]) => resolve(versions);
});

export const errorChart = createAction("ERROR_CHART", resolve => {
  return (err: Error) => resolve(err);
});

export const errorChartCatetories = createAction("ERROR_CHART_CATEGORIES", resolve => {
  return (err: Error) => resolve(err);
});

export const selectChartVersion = createAction("SELECT_CHART_VERSION", resolve => {
  return (chartVersion: IChartVersion, values?: string, schema?: JSONSchema4) =>
    resolve({ chartVersion, values, schema });
});

export const requestDeployedChartVersion = createAction("REQUEST_DEPLOYED_CHART_VERSION");

export const receiveDeployedChartVersion = createAction(
  "RECEIVE_DEPLOYED_CHART_VERSION",
  resolve => {
    return (chartVersion: IChartVersion, values?: string, schema?: JSONSchema4) =>
      resolve({ chartVersion, values, schema });
  },
);

export const resetChartVersion = createAction("RESET_CHART_VERSION");

export const resetRequestCharts = createAction("RESET_REQUEST_CHARTS");

export const selectReadme = createAction("SELECT_README", resolve => {
  return (readme: string) => resolve(readme);
});

export const errorReadme = createAction("ERROR_README", resolve => {
  return (message: string) => resolve(message);
});

const allActions = [
  requestCharts,
  requestChart,
  errorChart,
  errorChartCatetories,
  requestChartsCategories,
  receiveCharts,
  receiveChartCategories,
  receiveChartVersions,
  selectChartVersion,
  requestDeployedChartVersion,
  receiveDeployedChartVersion,
  resetChartVersion,
  resetRequestCharts,
  selectReadme,
  errorReadme,
];

export type ChartsAction = ActionType<typeof allActions[number]>;

export function fetchCharts(
  cluster: string,
  namespace: string,
  repos: string,
  page: number,
  size: number,
  query?: string,
): ThunkAction<Promise<void>, IStoreState, null, ChartsAction> {
  return async dispatch => {
    dispatch(requestCharts(page));
    try {
      const response = await Chart.fetchCharts(cluster, namespace, repos, page, size, query);
      dispatch(
        receiveCharts({
          items: response.data,
          page,
          totalPages: response.meta.totalPages,
        }),
      );
    } catch (e) {
      dispatch(errorChart(new FetchError(e.message)));
    }
  };
}

export function fetchChartCategories(
  cluster: string,
  namespace: string,
): ThunkAction<Promise<void>, IStoreState, null, ChartsAction> {
  return async dispatch => {
    dispatch(requestChartsCategories());
    try {
      const categories = await Chart.fetchChartCategories(cluster, namespace);
      if (categories) {
        dispatch(receiveChartCategories(categories));
      }
    } catch (e) {
      dispatch(errorChartCatetories(new FetchError(e.message)));
    }
  };
}

export function fetchChartVersions(
  cluster: string,
  namespace: string,
  id: string,
): ThunkAction<Promise<IChartVersion[]>, IStoreState, null, ChartsAction> {
  return async dispatch => {
    dispatch(requestCharts());
    try {
      const versions = await Chart.fetchChartVersions(cluster, namespace, id);
      if (versions) {
        dispatch(receiveChartVersions(versions));
      }
      return versions;
    } catch (e) {
      dispatch(errorChart(new FetchError(e.message)));
      return [];
    }
  };
}

async function getChart(cluster: string, namespace: string, id: string, version: string) {
  let values = "";
  let schema = {};
  const chartVersion = await Chart.getChartVersion(cluster, namespace, id, version);
  if (chartVersion) {
    try {
      values = await Chart.getValues(cluster, namespace, id, version);
      schema = await Chart.getSchema(cluster, namespace, id, version);
    } catch (e) {
      if (e.constructor !== NotFoundError) {
        throw e;
      }
    }
  }
  return { chartVersion, values, schema };
}

export function getChartVersion(
  cluster: string,
  namespace: string,
  id: string,
  version: string,
): ThunkAction<Promise<void>, IStoreState, null, ChartsAction> {
  return async dispatch => {
    try {
      dispatch(requestChart());
      const { chartVersion, values, schema } = await getChart(cluster, namespace, id, version);
      if (chartVersion) {
        dispatch(selectChartVersion(chartVersion, values, schema));
      }
    } catch (e) {
      dispatch(errorChart(new FetchError(e.message)));
    }
  };
}

export function getDeployedChartVersion(
  cluster: string,
  namespace: string,
  id: string,
  version: string,
): ThunkAction<Promise<void>, IStoreState, null, ChartsAction> {
  return async dispatch => {
    try {
      dispatch(requestDeployedChartVersion());
      const { chartVersion, values, schema } = await getChart(cluster, namespace, id, version);
      if (chartVersion) {
        dispatch(receiveDeployedChartVersion(chartVersion, values, schema));
      }
    } catch (e) {
      dispatch(errorChart(new FetchError(e.message)));
    }
  };
}

export function fetchChartVersionsAndSelectVersion(
  cluster: string,
  namespace: string,
  id: string,
  version?: string,
): ThunkAction<Promise<void>, IStoreState, null, ChartsAction> {
  return async dispatch => {
    const versions = (await dispatch(
      fetchChartVersions(cluster, namespace, id),
    )) as IChartVersion[];
    if (versions.length > 0) {
      let cv: IChartVersion = versions.sort((a, b) =>
        semver.compare(b.attributes.version, a.attributes.version),
      )[0];
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

export function getChartReadme(
  cluster: string,
  namespace: string,
  id: string,
  version: string,
): ThunkAction<Promise<void>, IStoreState, null, ChartsAction> {
  return async dispatch => {
    try {
      const readme = await Chart.getReadme(cluster, namespace, id, version);
      dispatch(selectReadme(readme));
    } catch (e) {
      dispatch(errorReadme(e.toString()));
    }
  };
}
