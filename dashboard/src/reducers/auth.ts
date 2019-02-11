import { getType } from "typesafe-actions";

import actions from "../actions";
import { AuthAction } from "../actions/auth";

export interface IAuthState {
  authenticated: boolean;
  authenticating: boolean;
  autoAuthenticating: boolean;
  autoAuthenticated: boolean;
  authenticationError?: string;
}

const initialState: IAuthState = {
  authenticated: !(localStorage.getItem("kubeapps_auth_token") === null),
  authenticating: false,
  autoAuthenticating: false,
  autoAuthenticated: false,
};

const authReducer = (state: IAuthState = initialState, action: AuthAction): IAuthState => {
  switch (action.type) {
    case getType(actions.auth.setAuthenticated):
      return {
        ...state,
        authenticated: action.payload,
        authenticating: false,
      };
    case getType(actions.auth.setAutoAuthenticated):
      return {
        ...state,
        authenticated: action.payload,
        autoAuthenticated: action.payload,
        authenticating: false,
        autoAuthenticating: false,
      };
    case getType(actions.auth.authenticating):
      return { ...state, authenticated: false, authenticating: true };
    case getType(actions.auth.autoAuthenticating):
      return { ...state, authenticated: false, autoAuthenticating: true };
    case getType(actions.auth.authenticationError):
      return {
        ...state,
        authenticated: false,
        authenticating: false,
        autoAuthenticating: false,
        autoAuthenticated: false,
        authenticationError: action.payload,
      };
    default:
  }
  return state;
};

export default authReducer;
