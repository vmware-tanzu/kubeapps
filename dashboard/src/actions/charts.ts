import { JSONSchema4 } from "json-schema";
import { Dispatch } from "redux";
import { ThunkAction } from "redux-thunk";
import { ActionType, createAction } from "typesafe-actions";

import Chart from "../shared/Chart";
import { IChart, IChartVersion, IStoreState, NotFoundError } from "../shared/types";

export const requestCharts = createAction("REQUEST_CHARTS");

export const receiveCharts = createAction("RECEIVE_CHARTS", resolve => {
  return (charts: IChart[]) => resolve(charts);
});

export const receiveChartVersions = createAction("RECEIVE_CHART_VERSIONS", resolve => {
  return (versions: IChartVersion[]) => resolve(versions);
});

export const errorChart = createAction("ERROR_CHART", resolve => {
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

export const selectReadme = createAction("SELECT_README", resolve => {
  return (readme: string) => resolve(readme);
});

export const errorReadme = createAction("ERROR_README", resolve => {
  return (message: string) => resolve(message);
});

const allActions = [
  requestCharts,
  errorChart,
  receiveCharts,
  receiveChartVersions,
  selectChartVersion,
  requestDeployedChartVersion,
  receiveDeployedChartVersion,
  resetChartVersion,
  selectReadme,
  errorReadme,
];

export type ChartsAction = ActionType<typeof allActions[number]>;

function dispatchError(dispatch: Dispatch, err: Error) {
  if (err.message.match("could not find")) {
    dispatch(errorChart(new NotFoundError(err.message)));
  } else {
    dispatch(errorChart(err));
  }
}

export function fetchCharts(
  repo: string,
): ThunkAction<Promise<void>, IStoreState, null, ChartsAction> {
  return async dispatch => {
    dispatch(requestCharts());
    try {
      const charts = await Chart.fetchCharts(repo);
      if (charts) {
        dispatch(receiveCharts(charts));
      }
    } catch (e) {
      dispatchError(dispatch, e);
    }
  };
}

export function fetchChartVersions(
  id: string,
): ThunkAction<Promise<IChartVersion[] | undefined>, IStoreState, null, ChartsAction> {
  return async dispatch => {
    dispatch(requestCharts());
    try {
      const versions = await Chart.fetchChartVersions(id);
      if (versions) {
        dispatch(receiveChartVersions(versions));
      }
      return versions;
    } catch (e) {
      dispatchError(dispatch, e);
      return;
    }
  };
}

async function getChart(id: string, version: string) {
  let values = "";
  let schema = {};
  const chartVersion = await Chart.getChartVersion(id, version);
  if (chartVersion) {
    try {
      values = await Chart.getValues(id, version);
      schema = await Chart.getSchema(id, version);
    } catch (e) {
      if (e.constructor !== NotFoundError) {
        throw e;
      }
    }
  }
  return { chartVersion, values, schema };
}

export function getChartVersion(
  id: string,
  version: string,
): ThunkAction<Promise<void>, IStoreState, null, ChartsAction> {
  return async dispatch => {
    try {
      dispatch(requestCharts());
      const { chartVersion, values, schema } = await getChart(id, version);
      if (chartVersion) {
        dispatch(selectChartVersion(chartVersion, values, schema));
      }
    } catch (e) {
      dispatchError(dispatch, e);
    }
  };
}

export function getDeployedChartVersion(
  id: string,
  version: string,
): ThunkAction<Promise<void>, IStoreState, null, ChartsAction> {
  return async dispatch => {
    try {
      dispatch(requestDeployedChartVersion());
      const { chartVersion, values, schema } = await getChart(id, version);
      if (chartVersion) {
        dispatch(receiveDeployedChartVersion(chartVersion, values, schema));
      }
    } catch (e) {
      dispatchError(dispatch, e);
    }
  };
}

export function fetchChartVersionsAndSelectVersion(
  id: string,
  version?: string,
): ThunkAction<Promise<void>, IStoreState, null, ChartsAction> {
  return async dispatch => {
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

export function getChartReadme(
  id: string,
  version: string,
): ThunkAction<Promise<void>, IStoreState, null, ChartsAction> {
  return async dispatch => {
    try {
      const readme = await Chart.getReadme(id, version);
      dispatch(selectReadme(readme));
    } catch (e) {
      dispatch(errorReadme(e.toString()));
    }
  };
}
