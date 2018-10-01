import { ThunkAction } from "redux-thunk";
import { ActionType, createAction } from "typesafe-actions";

import { Auth } from "../shared/Auth";
import { IStoreState } from "../shared/types";

export const setAuthenticated = createAction("SET_AUTHENTICATED", resolve => {
  return (authenticated: boolean) => resolve(authenticated);
});

export const authenticating = createAction("AUTHENTICATING");

export const authenticationError = createAction("AUTHENTICATION_ERROR", resolve => {
  return (errorMsg: string) => resolve(errorMsg);
});

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
