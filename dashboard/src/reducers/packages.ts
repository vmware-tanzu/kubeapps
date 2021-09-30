import { JSONSchemaType } from "ajv";
import { uniqBy } from "lodash";
import { IPackageState } from "shared/types";
import { getType } from "typesafe-actions";
import actions from "../actions";
import { PackagesAction } from "../actions/packages";
import { NamespaceAction } from "../actions/namespace";

export const initialState: IPackageState = {
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
  state: IPackageState["selected"],
  action: PackagesAction | NamespaceAction,
): IPackageState["selected"] => {
  switch (action.type) {
    case getType(actions.charts.receiveSelectedAvailablePackageDetail):
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
    case getType(actions.charts.resetPackageVersion):
      return initialState.selected;
    default:
  }
  return state;
};

const chartsReducer = (
  state: IPackageState = initialState,
  action: PackagesAction | NamespaceAction,
): IPackageState => {
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
    case getType(actions.charts.receiveSelectedAvailablePackageDetail):
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
          availablePackageDetail: action.payload.availablePackageDetail,
          schema: action.payload.availablePackageDetail.valuesSchema as any,
          values: action.payload.availablePackageDetail.defaultValues,
        },
      };
    case getType(actions.charts.resetAvailablePackageSummaries):
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
    case getType(actions.charts.resetPackageVersion):
    case getType(actions.charts.clearErrorPackage):
      return {
        ...state,
        selected: chartsSelectedReducer(state.selected, action),
      };
  }
  return state;
};

export default chartsReducer;
