import { getType } from 'typesafe-actions';

import * as actions from '../actions/index';
import { ChartState } from '../store/types';

const initialState: ChartState = {
  isFetching: false,
  selectedChart: null,
  selectedVersion: null,
  items: [],
};

function charts(state: ChartState = initialState, action: actions.RootAction): ChartState {
  switch (action.type) {
    case getType(actions.requestCharts):
      return {...state, isFetching: true};
    case getType(actions.receiveCharts):
      return {...state, isFetching: false, items: action.charts};
    case getType(actions.selectChart):
      return {...state, isFetching: false, selectedChart: action.chart};
    default:
  }
  return state;
}

export default {
  charts
};
