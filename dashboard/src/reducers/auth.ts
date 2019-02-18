import { getType } from "typesafe-actions";

import actions from "../actions";
import { AuthAction } from "../actions/auth";

export interface IAuthState {
  authenticated: boolean;
  authenticating: boolean;
  oidcAuthenticated: boolean;
  authenticationError?: string;
}

const initialState: IAuthState = {
  authenticated: !(localStorage.getItem("kubeapps_auth_token") === null),
  authenticating: false,
  oidcAuthenticated: localStorage.getItem("kubeapps_auth_token_oidc") === "true",
};

const authReducer = (state: IAuthState = initialState, action: AuthAction): IAuthState => {
  switch (action.type) {
    case getType(actions.auth.setAuthenticated):
      return {
        ...state,
        authenticated: action.payload.authenticated,
        oidcAuthenticated: action.payload.oidc,
        authenticating: false,
      };
    case getType(actions.auth.authenticating):
      return { ...state, authenticated: false, authenticating: true };
    case getType(actions.auth.authenticationError):
      return {
        ...state,
        authenticated: false,
        authenticating: false,
        oidcAuthenticated: false,
        authenticationError: action.payload,
      };
    default:
  }
  return state;
};

export default authReducer;
