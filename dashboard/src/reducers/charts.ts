import { uniqBy } from "lodash";
import { getType } from "typesafe-actions";

import actions from "../actions";
import { ChartsAction } from "../actions/charts";
import { NamespaceAction } from "../actions/namespace";
import { IChartState } from "../shared/types";

export const initialState: IChartState = {
  status: actions.charts.idleStatus,
  page: 1,
  size: 32,
  isFetching: false,
  items: [],
  selected: {
    versions: [],
  },
  deployed: {},
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
    case getType(actions.charts.receiveChartVersions):
      return {
        ...state,
        isFetching: false,
        status: actions.charts.idleStatus,
        selected: chartsSelectedReducer(state.selected, action),
      };
    case getType(actions.charts.selectChartVersion):
      return {
        ...state,
        isFetching: false,
        status: actions.charts.idleStatus,
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
        status: actions.charts.idleStatus,
        deployed: { ...state.deployed, ...action.payload },
      };
    case getType(actions.charts.resetChartVersion):
    case getType(actions.charts.selectReadme):
    case getType(actions.charts.errorReadme):
    case getType(actions.charts.requestCharts):
      return {
        ...state,
        isFetching: true,
        status: actions.charts.loadingStatus,
      };
    case getType(actions.charts.receiveCharts):
      return {
        ...state,
        items: uniqBy([...state.items, ...action.payload], "id"),
        page: state.page + 1,
        isFetching: false,
        status: actions.charts.idleStatus,
      };
    case getType(actions.charts.reachEnd):
      return {
        ...state,
        isFetching: false,
        status: actions.charts.finishedStatus,
      };
    case getType(actions.charts.errorChart):
      return {
        ...state,
        isFetching: false,
        status: actions.charts.errorStatus,
      };
    default:
  }
  return state;
};

export default chartsReducer;
