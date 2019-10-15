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
      return {
        ...state,
        error: undefined,
        readmeError: undefined,
        version: action.payload.chartVersion,
        values: action.payload.values ? action.payload.values : state.values,
        schema: action.payload.schema ? action.payload.schema : state.schema,
      };
    case getType(actions.charts.receiveChartVersions):
      return {
        ...state,
        error: undefined,
        versions: action.payload,
      };
    case getType(actions.charts.selectReadme):
      return { ...state, readme: action.payload, readmeError: undefined };
    case getType(actions.charts.errorChart):
      return { ...state, error: action.payload };
    case getType(actions.charts.errorReadme):
      return { ...state, readmeError: action.payload };
    case getType(actions.charts.resetChartVersion):
      return initialState.selected;
    default:
  }
  return state;
};

const chartsReducer = (state: IChartState = initialState, action: ChartsAction): IChartState => {
  switch (action.type) {
    case getType(actions.charts.requestCharts):
      return { ...state, isFetching: true };
    case getType(actions.charts.receiveCharts):
      return { ...state, isFetching: false, items: action.payload };
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
    case getType(actions.charts.resetChartVersion):
    case getType(actions.charts.selectReadme):
    case getType(actions.charts.errorReadme):
    case getType(actions.charts.errorChart):
      return { ...state, selected: chartsSelectedReducer(state.selected, action) };
    default:
  }
  return state;
};

export default chartsReducer;
