import { uniqBy } from "lodash";
import { getType } from "typesafe-actions";

import actions from "../actions";
import { ChartsAction } from "../actions/charts";
import { NamespaceAction } from "../actions/namespace";
import { IChartState } from "../shared/types";

export const initialState: IChartState = {
  isFetching: false,
  hasFinished: false,
  items: [],
  categories: [],
  selected: {
    versions: [],
  },
  deployed: {},
  page: 1,
  size: 100,
  records: new Map<number, boolean>().set(1, false),
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
      state.records.set(action.payload, false);
      return { ...state, isFetching: true, records: state.records };
    case getType(actions.charts.requestChartsCategories):
      return { ...state, isFetching: true };
    case getType(actions.charts.receiveCharts):
      // only update the records if the received page matches the requested one
      const areMatchingPages = action.payload.page === state.page;
      if (areMatchingPages) {
        state.records.set(state.page, true); // mark the current page as successfully retrieved
        state.records.set(state.page + 1, false); // mark the next page as pending (not undefined)
      }
      return {
        ...state,
        isFetching: false,
        hasFinished: action?.meta,
        items: areMatchingPages
          ? uniqBy([...state.items, ...action.payload.items], "id")
          : state.items, // if pages don't match, ignore payload to avoid duplicates
        page: action?.meta ? action.payload.page : action.payload.page + 1, // if action.meta==true, it's the last chunk
        records: state.records,
      };
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
    case getType(actions.charts.resetRequestCharts):
      state.records.clear();
      state.records.set(1, false);
      return {
        ...state,
        isFetching: false,
        hasFinished: false,
        items: [],
        page: 1,
        records: state.records,
      };
    case getType(actions.charts.resetChartVersion):
    case getType(actions.charts.selectReadme):
    case getType(actions.charts.errorReadme):
    case getType(actions.charts.errorChart):
      return {
        ...state,
        isFetching: false,
        hasFinished: false,
        page: 1,
        records: new Map<number, boolean>().set(1, false),
        items: [],
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
