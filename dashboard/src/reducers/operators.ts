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
  error?: Error;
  csvs: IClusterServiceVersion[];
}

const initialState: IOperatorsState = {
  isFetching: false,
  isOLMInstalled: false,
  operators: [],
  csvs: [],
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
      return { ...state, isFetching: false, error: action.payload };
    case getType(operators.requestCSVs):
      return { ...state, isFetching: true };
    case getType(operators.receiveCSVs):
      return { ...state, isFetching: false, csvs: action.payload };
    case getType(operators.errorCSVs):
      return { ...state, isFetching: false, error: action.payload };
    case LOCATION_CHANGE:
      return { ...state, error: undefined };
    case getType(actions.namespace.setNamespace):
      return { ...state, error: undefined };
    default:
      return { ...state };
  }
};

export default catalogReducer;
