import { getType } from 'typesafe-actions';

import * as actions from '../actions/index';
import { ChartState, Chart } from '../store/types';

const initialState: ChartState = {
  isFetching: false,
  items: [],
};

function charts(state: ChartState = initialState, action: actions.RootAction): ChartState {
  switch (action.type) {
    case getType(actions.requestCharts):
      return {...state, isFetching: true};
    case getType(actions.receiveCharts):
      return {...state, isFetching: false, items: <Chart[]> action.charts};
    default:
  }
  return state;
}

export default {
  charts
};
