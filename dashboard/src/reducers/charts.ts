import { uniqBy } from "lodash";
import { getType } from "typesafe-actions";

import actions from "../actions";
import { ChartsAction } from "../actions/charts";
import { NamespaceAction } from "../actions/namespace";
import { IChartState } from "../shared/types";

export const initialState: IChartState = {
  isFetching: false,
  hasFinishedFetching: false,
  items: [],
  categories: [],
  selected: {
    versions: [],
  },
  deployed: {},
  size: 20,
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
    case getType(actions.charts.clearErrorChart):
      return { ...state, error: undefined };
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
      return { ...state, isFetching: true };
    case getType(actions.charts.receiveCharts): {
      const isLastPage = action.payload.page >= action.payload.totalPages;
      return {
        ...state,
        isFetching: false,
        hasFinishedFetching: isLastPage,
        items: uniqBy([...state.items, ...action.payload.items], "id"),
      };
    }
    case getType(actions.charts.receiveChartCategories):
      return { ...state, categories: action.payload };
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
    case getType(actions.charts.resetRequestCharts):
      return {
        ...state,
        hasFinishedFetching: false,
        items: [],
      };
    case getType(actions.charts.resetChartVersion):
    case getType(actions.charts.selectReadme):
    case getType(actions.charts.errorReadme):
    case getType(actions.charts.errorChart):
      return {
        ...state,
        isFetching: false,
        hasFinishedFetching: false,
        items: state.items,
        selected: chartsSelectedReducer(state.selected, action),
      };
    case getType(actions.charts.clearErrorChart):
      return {
        ...state,
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
