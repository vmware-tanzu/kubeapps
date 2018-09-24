import { ThunkAction } from "redux-thunk";
import { ActionType, createActionDeprecated } from "typesafe-actions";

import { Auth } from "../shared/Auth";
import { IStoreState } from "../shared/types";

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

export function authenticate(
  token: string,
): ThunkAction<Promise<void>, IStoreState, null, AuthAction> {
  return async dispatch => {
    dispatch(authenticating());
    try {
      await Auth.validateToken(token);
      Auth.setAuthToken(token);
      dispatch(setAuthenticated(true));
    } catch (e) {
      dispatch(authenticationError(e.toString()));
    }
  };
}

export function logout(): ThunkAction<Promise<void>, IStoreState, null, AuthAction> {
  return async dispatch => {
    Auth.unsetAuthToken();
    dispatch(setAuthenticated(false));
  };
}
