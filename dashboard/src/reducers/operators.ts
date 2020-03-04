import { LocationChangeAction } from "connected-react-router";
import { getType } from "typesafe-actions";

import { OperatorAction } from "actions/operators";
import actions from "../actions";
import { NamespaceAction } from "../actions/namespace";

export interface IOperatorsState {
  isFetching: boolean;
  isOLMInstalled: boolean;
}

const initialState: IOperatorsState = {
  isFetching: false,
  isOLMInstalled: false,
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
    // TODO(andresmgot): Enable error cleanup when managing errors
    // case LOCATION_CHANGE:
    //   return { ...state, errors: {} };
    // case getType(actions.namespace.setNamespace):
    //   return { ...state, errors: {} };
    default:
      return { ...state };
  }
};

export default catalogReducer;
