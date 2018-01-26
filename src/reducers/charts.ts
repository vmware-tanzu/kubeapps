import { getType } from "typesafe-actions";

import actions from "../actions";
import { ChartsAction } from "../actions/charts";
import { IChartState } from "../shared/types";

const initialState: IChartState = {
  isFetching: false,
  items: [],
  selected: {
    versions: [],
  },
};

const chartsSelectedReducer = (
  state: IChartState["selected"],
  action: ChartsAction,
): IChartState["selected"] => {
  switch (action.type) {
    case getType(actions.charts.selectChartVersion):
      return { ...state, version: action.chartVersion };
    case getType(actions.charts.receiveChartVersions):
      return {
        ...state,
        versions: action.versions,
      };
    case getType(actions.charts.selectReadme):
      return { ...state, readme: action.readme };
    default:
  }
  return state;
};

const chartsReducer = (state: IChartState = initialState, action: ChartsAction): IChartState => {
  switch (action.type) {
    case getType(actions.charts.requestCharts):
      return { ...state, isFetching: true };
    case getType(actions.charts.receiveCharts):
      return { ...state, isFetching: false, items: action.charts };
    case getType(actions.charts.receiveChartVersions):
      return {
        ...state,
        isFetching: false,
        selected: chartsSelectedReducer(state.selected, action),
      };
    case getType(actions.charts.selectChartVersion):
      return {
        ...state,
        isFetching: false,
        selected: chartsSelectedReducer(state.selected, action),
      };
    case getType(actions.charts.selectReadme):
      return { ...state, selected: chartsSelectedReducer(state.selected, action) };
    default:
  }
  return state;
};

export default chartsReducer;
