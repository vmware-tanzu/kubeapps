import { getType } from "typesafe-actions";

import actions from "../actions";
import { ChartsAction } from "../actions/charts";
import { NamespaceAction } from "../actions/namespace";
import { IChartState } from "../shared/types";

export const initialState: IChartState = {
  status: actions.charts.idleStatus,
  isFetching: false,
  items: [],
  categories: [],
  selected: {
    versions: [],
  },
  deployed: {},
  search: {
    items: [],
    query: "",
  },
};

const chartsSelectedReducer = (
  state: IChartState["selected"],
  action: ChartsAction | NamespaceAction,
): IChartState["selected"] => {
  switch (action.type) {
    case getType(actions.charts.selectChartVersion):
      return {
        ...state,
        error: undefined,
        readmeError: undefined,
        version: action.payload.chartVersion,
        values: action.payload.values,
        schema: action.payload.schema,
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

const chartsReducer = (
  state: IChartState = initialState,
  action: ChartsAction | NamespaceAction,
): IChartState => {
  switch (action.type) {
    case getType(actions.charts.requestCharts):
      return { ...state, items: [], isFetching: true, status: actions.charts.loadingStatus };
    case getType(actions.charts.requestChartsCategories):
      return { ...state, isFetching: true };
    case getType(actions.charts.requestChartsSearch):
      return {
        ...state,
        isFetching: true,
        status:
          state.status === actions.charts.finishedStatus
            ? state.status
            : actions.charts.loadingStatus,
        search: { query: action.payload, items: [] },
      };
    case getType(actions.charts.receiveCharts):
      return {
        ...state,
        isFetching: false,
        items: action.payload,
        status:
          state.status === actions.charts.finishedStatus ? state.status : actions.charts.idleStatus,
      };
    case getType(actions.charts.receiveChartsSearch):
      if (action.meta === state.search.query) {
        return {
          ...state,
          search: { items: action.payload, query: action.meta },
          isFetching: false,
          status: actions.charts.finishedStatus,
        };
      } else {
        return state;
      }
    case getType(actions.charts.receiveChartCategories):
      return { ...state, isFetching: false, categories: action.payload };
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
    case getType(actions.charts.requestDeployedChartVersion):
      return {
        ...state,
        deployed: {},
      };
    case getType(actions.charts.receiveDeployedChartVersion):
      return {
        ...state,
        isFetching: false,
        deployed: { ...state.deployed, ...action.payload },
      };
    case getType(actions.charts.resetChartsSearch):
      return { ...state, search: { items: [], query: "" }, status: actions.charts.idleStatus };
    case getType(actions.charts.resetChartVersion):
    case getType(actions.charts.selectReadme):
    case getType(actions.charts.errorReadme):
    case getType(actions.charts.errorChart):
      return {
        ...state,
        isFetching: false,
        status: actions.charts.errorStatus,
        selected: chartsSelectedReducer(state.selected, action),
      };
    case getType(actions.charts.errorChartCatetories):
      return {
        ...state,
        isFetching: false,
        categories: [],
      };
    default:
  }
  return state;
};

export default chartsReducer;
