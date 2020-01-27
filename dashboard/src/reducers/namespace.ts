import { LOCATION_CHANGE, LocationChangeAction } from "connected-react-router";
import { getType } from "typesafe-actions";

import actions from "../actions";
import { AuthAction } from "../actions/auth";
import { NamespaceAction } from "../actions/namespace";
import { Auth } from "../shared/Auth";

export interface INamespaceState {
  current: string;
  namespaces: string[];
  error?: { action: string; errorMsg: string };
}

const getInitialState: () => INamespaceState = (): INamespaceState => {
  const token = Auth.getAuthToken() || "";
  return {
    current: Auth.defaultNamespaceFromToken(token),
    namespaces: [],
  };
};
const initialState: INamespaceState = getInitialState();

const namespaceReducer = (
  state: INamespaceState = initialState,
  action: NamespaceAction | LocationChangeAction | AuthAction,
): INamespaceState => {
  switch (action.type) {
    case getType(actions.namespace.receiveNamespaces):
      return { ...state, namespaces: action.payload };
    case getType(actions.namespace.setNamespace):
      return { ...state, current: action.payload, error: undefined };
    case getType(actions.namespace.errorNamespaces):
      return {
        ...state,
        error: { action: action.payload.op, errorMsg: action.payload.err.message },
      };
    case getType(actions.namespace.clearNamespaces):
      return { ...initialState };
    case LOCATION_CHANGE:
      const pathname = action.payload.location.pathname;
      // looks for /ns/:namespace in URL
      const matches = pathname.match(/\/ns\/([^/]*)/);
      if (matches) {
        return { ...state, current: matches[1] };
      }
      break;
    case getType(actions.auth.setAuthenticated):
      // Only when a user is authenticated to we set the current namespace from
      // the auth default namespace.
      if (action.payload.authenticated) {
        return { ...state, current: action.payload.defaultNamespace };
      }
    default:
  }
  return state;
};

export default namespaceReducer;
