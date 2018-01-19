import { getType } from "typesafe-actions";

import actions from "../actions";
import { ChartsAction } from "../actions/charts";
import { IChartState } from "../shared/types";

const initialState: IChartState = {
  isFetching: false,
  items: [],
  selectedChart: null,
  selectedVersion: null,
};

const chartsReducer = (state: IChartState = initialState, action: ChartsAction): IChartState => {
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
};

export default chartsReducer;
