import { LOCATION_CHANGE, LocationChangeAction } from "connected-react-router";
import { getType } from "typesafe-actions";

import { OperatorAction } from "actions/operators";
import actions from "../actions";
import { NamespaceAction } from "../actions/namespace";
import { IClusterServiceVersion, IPackageManifest, IResource } from "../shared/types";

export interface IOperatorsStateError {
  fetch?: Error;
  create?: Error;
  delete?: Error;
  update?: Error;
}

export interface IOperatorsState {
  isFetching: boolean;
  isFetchingElem: {
    OLM: boolean;
    operator: boolean;
    csv: boolean;
    resource: boolean;
  };
  isOLMInstalled: boolean;
  operators: IResource[];
  operator?: IPackageManifest;
  errors: {
    operator: IOperatorsStateError;
    csv: IOperatorsStateError;
    resource: IOperatorsStateError;
  };
  csvs: IClusterServiceVersion[];
  csv?: IClusterServiceVersion;
  resources: IResource[];
  resource?: IResource;
}

export const operatorsInitialState: IOperatorsState = {
  isFetching: false,
  isFetchingElem: {
    OLM: false,
    operator: false,
    csv: false,
    resource: false,
  },
  isOLMInstalled: false,
  operators: [],
  csvs: [],
  errors: {
    operator: {},
    csv: {},
    resource: {},
  },
  resources: [],
};

function isFetching(state: IOperatorsState, item: string, fetching: boolean) {
  const composedIsFetching = {
    ...state.isFetchingElem,
    [item]: fetching,
  };
  return {
    isFetching: Object.values(composedIsFetching).some(v => v),
    isFetchingElem: composedIsFetching,
  };
}

const catalogReducer = (
  state: IOperatorsState = operatorsInitialState,
  action: OperatorAction | LocationChangeAction | NamespaceAction,
): IOperatorsState => {
  const { operators } = actions;
  switch (action.type) {
    case getType(operators.checkingOLM):
      return { ...state, ...isFetching(state, "OLM", true) };
    case getType(operators.OLMInstalled):
      return { ...state, isOLMInstalled: true, ...isFetching(state, "OLM", false) };
    case getType(operators.errorOLMCheck):
      return {
        ...state,
        ...isFetching(state, "OLM", false),
        errors: { ...state.errors, operator: { fetch: action.payload } },
      };
    case getType(operators.requestOperators):
      return { ...state, ...isFetching(state, "operator", true) };
    case getType(operators.receiveOperators):
      return {
        ...state,
        ...isFetching(state, "operator", false),
        operators: action.payload,
      };
    case getType(operators.requestOperator):
      return { ...state, ...isFetching(state, "operator", true) };
    case getType(operators.receiveOperator):
      return {
        ...state,
        ...isFetching(state, "operator", false),
        operator: action.payload,
      };
    case getType(operators.errorOperators):
      return {
        ...state,
        ...isFetching(state, "operator", false),
        operator: undefined,
        errors: { ...state.errors, operator: { fetch: action.payload } },
      };
    case getType(operators.requestCSVs):
      return { ...state, ...isFetching(state, "csv", true) };
    case getType(operators.receiveCSVs):
      return { ...state, ...isFetching(state, "csv", false), csvs: action.payload };
    case getType(operators.requestCSV):
      return { ...state, ...isFetching(state, "csv", true) };
    case getType(operators.receiveCSV):
      return { ...state, ...isFetching(state, "csv", false), csv: action.payload };
    case getType(operators.errorCSVs):
      return {
        ...state,
        ...isFetching(state, "csv", false),
        csv: undefined,
        errors: { ...state.errors, csv: { fetch: action.payload } },
      };
    case getType(operators.creatingResource):
      return { ...state, ...isFetching(state, "resource", true) };
    case getType(operators.resourceCreated):
      return { ...state, ...isFetching(state, "resource", false) };
    case getType(operators.updatingResource):
      return { ...state, ...isFetching(state, "resource", true) };
    case getType(operators.resourceUpdated):
      return { ...state, ...isFetching(state, "resource", false) };
    case getType(operators.deletingResource):
      return { ...state, ...isFetching(state, "resource", true) };
    case getType(operators.resourceDeleted):
      return { ...state, ...isFetching(state, "resource", false) };
    case getType(operators.errorResourceDelete):
      return {
        ...state,
        ...isFetching(state, "resource", false),
        errors: { ...state.errors, resource: { delete: action.payload } },
      };
    case getType(operators.errorResourceCreate):
      return {
        ...state,
        ...isFetching(state, "resource", false),
        errors: { ...state.errors, resource: { create: action.payload } },
      };
    case getType(operators.errorResourceUpdate):
      return {
        ...state,
        ...isFetching(state, "resource", false),
        errors: { ...state.errors, resource: { update: action.payload } },
      };
    case getType(operators.requestCustomResources):
      return { ...state, ...isFetching(state, "resource", true) };
    case getType(operators.receiveCustomResources):
      return {
        ...state,
        ...isFetching(state, "resource", false),
        resources: action.payload,
      };
    case getType(operators.errorCustomResource):
      return {
        ...state,
        ...isFetching(state, "resource", false),
        resource: undefined,
        errors: { ...state.errors, resource: { fetch: action.payload } },
      };
    case getType(operators.requestCustomResource):
      return { ...state, ...isFetching(state, "resource", true) };
    case getType(operators.receiveCustomResource):
      return {
        ...state,
        ...isFetching(state, "resource", false),
        resource: action.payload,
      };
    case getType(operators.creatingOperator):
      return { ...state, ...isFetching(state, "operator", true) };
    case getType(operators.operatorCreated):
      return { ...state, ...isFetching(state, "operator", false) };
    case getType(operators.errorOperatorCreate):
      return {
        ...state,
        ...isFetching(state, "operator", false),
        errors: { ...state.errors, operator: { create: action.payload } },
      };
    case LOCATION_CHANGE:
      return {
        ...state,
        isOLMInstalled: state.isOLMInstalled,
        errors: { operator: {}, csv: {}, resource: {} },
      };
    case getType(actions.namespace.setNamespace):
      return { ...operatorsInitialState, isOLMInstalled: state.isOLMInstalled };
    default:
      return { ...state };
  }
};

export default catalogReducer;
