import { getType } from "typesafe-actions";

import actions from "../actions";
import { AuthAction } from "../actions/auth";
import { Auth } from "../shared/Auth";
import { definedNamespaces } from "../shared/Namespace";

export interface IAuthState {
  sessionExpired: boolean;
  authenticated: boolean;
  authenticating: boolean;
  oidcAuthenticated: boolean;
  authenticationError?: string;
  defaultNamespace: string;
}

const getInitialState: () => IAuthState = (): IAuthState => {
  const token = Auth.getAuthToken() || "";
  return {
    sessionExpired: false,
    authenticated: token !== "",
    authenticating: false,
    oidcAuthenticated: Auth.usingOIDCToken(),
    defaultNamespace: Auth.defaultNamespaceFromToken(token),
  };
};
const initialState: IAuthState = getInitialState();

const authReducer = (state: IAuthState = initialState, action: AuthAction): IAuthState => {
  switch (action.type) {
    case getType(actions.auth.setAuthenticated):
      return {
        ...state,
        authenticated: action.payload.authenticated,
        oidcAuthenticated: action.payload.oidc,
        authenticating: false,
        defaultNamespace: action.payload.defaultNamespace || definedNamespaces.all,
      };
    case getType(actions.auth.authenticating):
      return { ...state, authenticated: false, authenticating: true };
    case getType(actions.auth.authenticationError):
      return {
        ...state,
        authenticationError: action.payload,
        authenticating: false,
      };
    case getType(actions.auth.setSessionExpired):
      return {
        ...state,
        sessionExpired: action.payload.sessionExpired,
      };
    default:
  }
  return state;
};

export default authReducer;
