import { Dispatch } from "redux";
import { createAction, getReturnOfExpression } from "typesafe-actions";

import { Auth } from "../shared/Auth";

export const setAuthenticated = createAction("SET_AUTHENTICATED", (authenticated: boolean) => ({
  authenticated,
  type: "SET_AUTHENTICATED",
}));

export const authenticating = createAction("AUTHENTICATING", () => ({
  type: "AUTHENTICATING",
}));

export const authenticationError = createAction("AUTHENTICATION_ERROR", (errorMsg: string) => ({
  errorMsg,
  type: "AUTHENTICATION_ERROR",
}));

const allActions = [setAuthenticated, authenticating, authenticationError].map(
  getReturnOfExpression,
);
export type AuthAction = typeof allActions[number];

export function authenticate(token: string) {
  return async (dispatch: Dispatch) => {
    dispatch(authenticating());
    try {
      await Auth.validateToken(token);
      Auth.setAuthToken(token);
      return dispatch(setAuthenticated(true));
    } catch (e) {
      return dispatch(authenticationError(e.toString()));
    }
  };
}

export function logout() {
  return async (dispatch: Dispatch) => {
    Auth.unsetAuthToken();
    return dispatch(setAuthenticated(false));
  };
}
