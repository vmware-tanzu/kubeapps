import { JSONSchemaType } from "ajv";
import { uniqBy } from "lodash";
import { IChartState } from "shared/types";
import { getType } from "typesafe-actions";
import actions from "../actions";
import { ChartsAction } from "../actions/packages";
import { NamespaceAction } from "../actions/namespace";

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
        availablePackageDetail: action.payload.selectedPackage,
        pkgVersion: action.payload.selectedPackage.version?.pkgVersion,
        appVersion: action.payload.selectedPackage.version?.appVersion,
        readme: action.payload.selectedPackage.readme,
        values: action.payload.selectedPackage.defaultValues,
        schema:
          action.payload.selectedPackage.valuesSchema !== ""
            ? (JSON.parse(action.payload.selectedPackage.valuesSchema) as JSONSchemaType<any>)
            : ({} as JSONSchemaType<any>),
      };
    case getType(actions.charts.receiveAvailablePackageVersions):
      return {
        ...state,
        error: undefined,
        versions: action.payload.packageAppVersions,
      };
    case getType(actions.charts.errorPackage):
      return { ...state, error: action.payload };
    case getType(actions.charts.clearErrorPackage):
      return { ...state, error: undefined };
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
    case getType(actions.charts.requestAvailablePackageSummaries):
      return { ...state, isFetching: true };
    case getType(actions.charts.receiveAvailablePackageSummaries): {
      const isLastPage =
        action.payload.page >= parseInt(action.payload.response.nextPageToken) ||
        action.payload.response.nextPageToken === "";
      return {
        ...state,
        isFetching: false,
        hasFinishedFetching: isLastPage,
        categories: action.payload.response.categories,
        items: uniqBy(
          [...state.items, ...action.payload.response.availablePackageSummaries],
          "availablePackageRef.identifier",
        ),
      };
    }
    case getType(actions.charts.receiveAvailablePackageVersions):
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
    case getType(actions.charts.requestDeployedAvailablePackageDetail):
      return {
        ...state,
        deployed: {},
      };
    case getType(actions.charts.receiveDeployedAvailablePackageDetail):
      return {
        ...state,
        isFetching: false,
        deployed: {
          chartVersion: action.payload.chartVersion,
          schema: action.payload.schema as any,
          values: action.payload.values,
        },
      };
    case getType(actions.charts.resetRequestCharts):
      return {
        ...state,
        hasFinishedFetching: false,
        items: [],
      };
    case getType(actions.charts.errorPackage):
      return {
        ...state,
        isFetching: false,
        hasFinishedFetching: false,
        items: state.items,
        selected: chartsSelectedReducer(state.selected, action),
      };
    case getType(actions.charts.resetChartVersion):
    case getType(actions.charts.clearErrorPackage):
      return {
        ...state,
        selected: chartsSelectedReducer(state.selected, action),
      };
  }
  return state;
};

export default chartsReducer;
