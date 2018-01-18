import { getType } from 'typesafe-actions';

import actions from '../actions';
import { ChartsAction } from '../actions/charts';
import { ChartState } from '../store/types';

const initialState: ChartState = {
  isFetching: false,
  selectedChart: null,
  selectedVersion: null,
  items: [],
};

function charts(state: ChartState = initialState, action: ChartsAction): ChartState {
  switch (action.type) {
    case getType(actions.charts.requestCharts):
      return { ...state, isFetching: true };
    case getType(actions.charts.receiveCharts):
      return { ...state, isFetching: false, items: action.charts };
    case getType(actions.charts.selectChart):
      return { ...state, isFetching: false, selectedChart: action.chart };
    default:
  }
  return state;
}

export default {
  charts
};
