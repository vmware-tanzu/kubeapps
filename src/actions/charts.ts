import { Dispatch } from 'redux';
import { createAction, getReturnOfExpression } from 'typesafe-actions';

import { StoreState, Chart } from '../shared/types';
import * as url from '../shared/url';

export const requestCharts = createAction('REQUEST_CHARTS');
export const receiveCharts = createAction('RECEIVE_CHARTS', (charts: Chart[]) => ({
  type: 'RECEIVE_CHARTS',
  charts
}));
export const selectChart = createAction('SELECT_CHART', (chart: Chart) => ({
  type: 'SELECT_CHART',
  chart
}));

const allActions = [requestCharts, receiveCharts, selectChart].map(getReturnOfExpression);
export type ChartsAction = typeof allActions[number];

export function fetchCharts(repo: string) {
  return (dispatch: Dispatch<StoreState>): Promise<{}> => {
    dispatch(requestCharts());
    return fetch(url.api.charts.list(repo))
      .then(response => response.json())
      .then(json => dispatch(receiveCharts(json.data)));
  };
}

export function getChart(id: string) {
  return (dispatch: Dispatch<StoreState>): Promise<{}> => {
    dispatch(requestCharts());
    return fetch(url.api.charts.get(id))
      .then(response => response.json())
      .then(json => dispatch(selectChart(json.data)));
  };
}

export function deployChart(chart: Chart, releaseName: string, namespace: string) {
  return (dispatch: Dispatch<StoreState>): Promise<{}> => {
    return fetch(url.api.helmreleases.create(namespace), {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        apiVersion: 'helm.bitnami.com/v1',
        kind: 'HelmRelease',
        metadata: {
          name: releaseName,
        },
        spec: {
          repoUrl: chart.attributes.repo.url,
          chartName: chart.attributes.name,
          version: chart.relationships.latestChartVersion.data.version,
        }
      }),
    }).then(response => response.json())
      .then(json => {
        if (json.status === 'Failure') {
          throw new Error(json.message);
        }
        return json;
      });
  };
}
