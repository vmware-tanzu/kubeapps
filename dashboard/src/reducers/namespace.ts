import { LOCATION_CHANGE, LocationChangeAction } from "connected-react-router";
import { getType } from "typesafe-actions";

import actions from "../actions";
import { NamespaceAction } from "../actions/namespace";

export interface INamespaceState {
  current: string;
  namespaces: string[];
  errorMsg?: string;
}

const initialState: INamespaceState = {
  current: "",
  namespaces: [],
};

const namespaceReducer = (
  state: INamespaceState = initialState,
  action: NamespaceAction | LocationChangeAction,
): INamespaceState => {
  switch (action.type) {
    case getType(actions.namespace.receiveNamespaces):
      return { ...state, namespaces: action.payload };
    case getType(actions.namespace.setNamespace):
      return { ...state, current: action.payload };
    case getType(actions.namespace.errorNamespaces):
      return { ...state, errorMsg: action.payload.err.message };
    case getType(actions.namespace.clearNamespaces):
      return { ...initialState };
    case LOCATION_CHANGE:
      const pathname = action.payload.location.pathname;
      // looks for /ns/:namespace in URL
      const matches = pathname.match(/\/ns\/([^/]*)/);
      if (matches) {
        return { ...state, current: matches[1] };
      }
    default:
  }
  return state;
};

export default namespaceReducer;
