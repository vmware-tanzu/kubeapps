import { Dispatch } from "redux";
import { ThunkAction } from "redux-thunk";
import * as semver from "semver";
import { ActionType, createAction } from "typesafe-actions";

import Chart from "../shared/Chart";
import {
  IChart,
  IChartUpdate,
  IChartUpdatesList,
  IChartVersion,
  IStoreState,
  NotFoundError,
} from "../shared/types";
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

export const receiveChartUpdates = createAction("RECEIVE_CHART_UPDATES", resolve => {
  return (chartUpdates: IChartUpdatesList) => resolve(chartUpdates);
});

export const errorChartUpdates = createAction("ERROR_CHART_UPDATES", resolve => {
  return (err: Error) => resolve(err);
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
  receiveChartUpdates,
  errorChartUpdates,
];

export type ChartsAction = ActionType<typeof allActions[number]>;

async function httpGet(dispatch: Dispatch, targetURL: string): Promise<any> {
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

export function listChartsWithFilters(
  name: string,
  version: string,
  appVersion: string,
): ThunkAction<Promise<void>, IStoreState, null, ChartsAction> {
  return async dispatch => {
    try {
      const chartUpdates = await Chart.listWithFilters(name, version, appVersion);
      const updates: IChartUpdate[] = [];
      chartUpdates.forEach(c => {
        // semver.compare returns -1 if v2 is bigger than v1
        if (semver.compare(version, c.relationships.latestChartVersion.data.version) < 0) {
          updates.push({
            latestVersion: c.relationships.latestChartVersion.data.version,
            repository: c.attributes.repo,
          });
        }
      });
      dispatch(receiveChartUpdates({ name, updates }));
    } catch (e) {
      errorChartUpdates(e);
    }
  };
}
