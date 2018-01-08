import { Dispatch } from 'redux';
import { createAction, getReturnOfExpression } from 'typesafe-actions';

import { StoreState } from '../store/types';

export const requestCharts = createAction('REQUEST_CHARTS');
export const receiveCharts = createAction('RECEIVE_CHARTS', (charts: Array<{}>) => ({
  type: 'RECEIVE_CHARTS',
  charts
}));

export function fetchCharts() {
  return (dispatch: Dispatch<StoreState>): Promise<{}> => {
    dispatch(requestCharts());
    return fetch(`/api/chartsvc/v1/charts`)
      .then(response => response.json())
      .then(json => dispatch(receiveCharts(json.data)));
  };
}

const allActions = [requestCharts, receiveCharts].map(getReturnOfExpression);
export type RootAction = typeof allActions[number];
