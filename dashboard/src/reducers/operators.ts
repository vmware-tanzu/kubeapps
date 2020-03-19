import { LOCATION_CHANGE, LocationChangeAction } from "connected-react-router";
import { getType } from "typesafe-actions";

import { OperatorAction } from "actions/operators";
import actions from "../actions";
import { NamespaceAction } from "../actions/namespace";
import { IClusterServiceVersion, IPackageManifest, IResource } from "../shared/types";

export interface IOperatorsState {
  isFetching: boolean;
  isOLMInstalled: boolean;
  operators: IResource[];
  operator?: IPackageManifest;
  errors: {
    fetch?: Error;
    create?: Error;
    delete?: Error;
  };
  csvs: IClusterServiceVersion[];
  csv?: IClusterServiceVersion;
  resources: IResource[];
  resource?: IResource;
}

const initialState: IOperatorsState = {
  isFetching: false,
  isOLMInstalled: false,
  operators: [],
  csvs: [],
  errors: {},
  resources: [],
};

const catalogReducer = (
  state: IOperatorsState = initialState,
  action: OperatorAction | LocationChangeAction | NamespaceAction,
): IOperatorsState => {
  const { operators } = actions;
  switch (action.type) {
    case getType(operators.checkingOLM):
      return { ...state, isFetching: true };
    case getType(operators.OLMInstalled):
      return { ...state, isOLMInstalled: true, isFetching: false };
    case getType(operators.OLMNotInstalled):
      return { ...state, isOLMInstalled: false, isFetching: false };
    case getType(operators.requestOperators):
      return { ...state, isFetching: true };
    case getType(operators.receiveOperators):
      return { ...state, isFetching: false, operators: action.payload };
    case getType(operators.requestOperator):
      return { ...state, isFetching: true };
    case getType(operators.receiveOperator):
      return { ...state, isFetching: false, operator: action.payload };
    case getType(operators.errorOperators):
      return { ...state, isFetching: false, errors: { fetch: action.payload } };
    case getType(operators.requestCSVs):
      return { ...state, isFetching: true };
    case getType(operators.receiveCSVs):
      return { ...state, isFetching: false, csvs: action.payload };
    case getType(operators.requestCSV):
      return { ...state, isFetching: true };
    case getType(operators.receiveCSV):
      return { ...state, isFetching: false, csv: action.payload };
    case getType(operators.errorCSVs):
      return { ...state, isFetching: false, errors: { fetch: action.payload } };
    case getType(operators.creatingResource):
      return { ...state, isFetching: true };
    case getType(operators.resourceCreated):
      return { ...state, isFetching: false };
    case getType(operators.deletingResource):
      return { ...state, isFetching: true };
    case getType(operators.resourceDeleted):
      return { ...state, isFetching: false };
    case getType(operators.errorResourceDelete):
      return { ...state, isFetching: false, errors: { delete: action.payload } };
    case getType(operators.errorResourceCreate):
      return { ...state, isFetching: false, errors: { create: action.payload } };
    case getType(operators.requestCustomResources):
      return { ...state, isFetching: true };
    case getType(operators.receiveCustomResources):
      return { ...state, isFetching: false, resources: action.payload };
    case getType(operators.errorCustomResource):
      return { ...state, isFetching: false, errors: { fetch: action.payload } };
    case getType(operators.requestCustomResource):
      return { ...state, isFetching: true };
    case getType(operators.receiveCustomResource):
      return { ...state, isFetching: false, resource: action.payload };
    case LOCATION_CHANGE:
      return { ...state, errors: {} };
    case getType(actions.namespace.setNamespace):
      return { ...state, errors: {} };
    default:
      return { ...state };
  }
};

export default catalogReducer;
