import { Dispatch } from "redux";
import { ThunkAction } from "redux-thunk";
import { ActionType, createAction } from "typesafe-actions";

import { axios } from "../shared/AxiosInstance";
import Chart from "../shared/Chart";
import { IChart, IChartVersion, IStoreState, NotFoundError } from "../shared/types";
import * as url from "../shared/url";

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
  return (chartVersion: IChartVersion) => resolve(chartVersion);
});

export const resetChartVersion = createAction("RESET_CHART_VERSION");

export const selectReadme = createAction("SELECT_README", resolve => {
  return (readme: string) => resolve(readme);
});

export const errorReadme = createAction("ERROR_README", resolve => {
  return (message: string) => resolve(message);
});

export const selectValues = createAction("SELECT_VALUES", resolve => {
  return (values: string) => resolve(values);
});

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
];

export type ChartsAction = ActionType<typeof allActions[number]>;

async function httpGet(dispatch: Dispatch, targetURL: string): Promise<any> {
  try {
    const response = await axios.get(targetURL);
    // If non-2XX response (not ok)
    if (response.status < 200 || response.status > 299) {
      const error = response.data || response.statusText;
      if (response.status === 404) {
        dispatch(errorChart(new NotFoundError(error)));
      } else {
        dispatch(errorChart(new Error(error)));
      }
    } else {
      return response.data.data;
    }
  } catch (e) {
    dispatch(errorChart(e));
  }
}

export function fetchCharts(
  repo: string,
): ThunkAction<Promise<void>, IStoreState, null, ChartsAction> {
  return async dispatch => {
    dispatch(requestCharts());
    const response = await httpGet(dispatch, url.api.charts.list(repo));
    if (response) {
      dispatch(receiveCharts(response));
    }
  };
}

export function fetchChartVersions(
  id: string,
): ThunkAction<Promise<IChartVersion[]>, IStoreState, null, ChartsAction> {
  return async dispatch => {
    dispatch(requestCharts());
    const response = await httpGet(dispatch, url.api.charts.listVersions(id));
    if (response) {
      dispatch(receiveChartVersions(response));
    }
    return response;
  };
}

export function getChartVersion(
  id: string,
  version: string,
): ThunkAction<Promise<void>, IStoreState, null, ChartsAction> {
  return async dispatch => {
    dispatch(requestCharts());
    const response = await httpGet(dispatch, url.api.charts.getVersion(id, version));
    if (response) {
      dispatch(selectChartVersion(response));
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

export function getChartValues(
  id: string,
  version: string,
): ThunkAction<Promise<void>, IStoreState, null, ChartsAction> {
  return async dispatch => {
    try {
      const values = await Chart.getValues(id, version);
      dispatch(selectValues(values));
    } catch (e) {
      dispatch(selectValues(""));
    }
  };
}
