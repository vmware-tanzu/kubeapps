import { uniqBy } from "lodash";
import { getType } from "typesafe-actions";

import actions from "../actions";
import { ChartsAction } from "../actions/charts";
import { NamespaceAction } from "../actions/namespace";
import { IChartState } from "../shared/types";

export const initialState: IChartState = {
  status: actions.charts.unstartedStatus,
  nextPage: 1,
  page: 1,
  size: 100,
  isFetching: false,
  items: [],
  search: {
    items: [],
    query: "",
  },
  categories: [],
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
        selected: chartsSelectedReducer(state.selected, action),
      };
    case getType(actions.charts.receiveChartCategories):
      return {
        ...state,
        isFetching: false,
        categories: action.payload,
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
    case getType(actions.charts.resetChartVersion):
      return {
        ...state,
        isFetching: false,
        selected: chartsSelectedReducer(state.selected, action),
      };
    case getType(actions.charts.selectReadme):
      return {
        ...state,
        isFetching: false,
        selected: chartsSelectedReducer(state.selected, action),
      };
    case getType(actions.charts.errorReadme):
      return {
        ...state,
        isFetching: false,
        selected: chartsSelectedReducer(state.selected, action),
      };
    case getType(actions.charts.requestCharts):
      return {
        ...state,
        isFetching: true,
        nextPage: action.payload + 1 > state.nextPage ? action.payload + 1 : state.nextPage,
        // status: state.status === actions.charts.finishedStatus ? state.status : actions.charts.loadingStatus,
        status: actions.charts.loadingStatus,
      };
    case getType(actions.charts.requestChartsCategories):
      return {
        ...state,
        isFetching: true,
      };
    case getType(actions.charts.requestChartsVersions):
      return {
        // ...state,
        // selected: initialState.selected,
        ...initialState,
        isFetching: true,
      };
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
        items: uniqBy([...state.items, ...action.payload], "id"), // TODO(agamez): add canceling request features to avoid this workaround
        page: state.page === state.nextPage ? state.page + 1 : state.nextPage,
        isFetching: false,
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
        selected: chartsSelectedReducer(state.selected, action),
      };
    case getType(actions.charts.resetChartsSearch):
      return {
        ...state,
        search: { items: [], query: "" },
        status: actions.charts.idleStatus,
      };
    case getType(actions.charts.resetPaginaton):
      return {
        ...state,
        nextPage: initialState.nextPage,
        page: initialState.page,
        items: initialState.items,
        isFetching: false,
        status: actions.charts.idleStatus,
      };
    default:
  }
  return state;
};

export default chartsReducer;
