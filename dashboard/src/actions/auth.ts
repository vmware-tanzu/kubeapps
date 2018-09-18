import { Dispatch } from "redux";
import { ActionType, createActionDeprecated } from "typesafe-actions";

import { Auth } from "../shared/Auth";

export const setAuthenticated = createActionDeprecated(
  "SET_AUTHENTICATED",
  (authenticated: boolean) => ({
    authenticated,
    type: "SET_AUTHENTICATED",
  }),
);

export const authenticating = createActionDeprecated("AUTHENTICATING", () => ({
  type: "AUTHENTICATING",
}));

export const authenticationError = createActionDeprecated(
  "AUTHENTICATION_ERROR",
  (errorMsg: string) => ({
    errorMsg,
    type: "AUTHENTICATION_ERROR",
  }),
);

const allActions = [setAuthenticated, authenticating, authenticationError];

export type AuthAction = ActionType<typeof allActions[number]>;

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
