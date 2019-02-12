import { ThunkAction } from "redux-thunk";
import { ActionType, createAction } from "typesafe-actions";

import { Auth } from "../shared/Auth";
import { IStoreState } from "../shared/types";

export const setAuthenticated = createAction("SET_AUTHENTICATED", resolve => {
  return (authenticated: boolean, withToken?: boolean) => resolve({ authenticated, withToken });
});

export const authenticating = createAction("AUTHENTICATING");

export const authenticationError = createAction("AUTHENTICATION_ERROR", resolve => {
  return (errorMsg: string) => resolve(errorMsg);
});

export const checkingOIDCToken = createAction("CHECKING_OIDC_TOKEN");

const allActions = [setAuthenticated, authenticating, authenticationError, checkingOIDCToken];

export type AuthAction = ActionType<typeof allActions[number]>;

export function authenticate(
  token?: string,
): ThunkAction<Promise<void>, IStoreState, null, AuthAction> {
  return async dispatch => {
    dispatch(authenticating());
    try {
      await Auth.validate(token);
      if (token) {
        Auth.setAuthToken(token);
      }
      dispatch(setAuthenticated(true, !!token));
    } catch (e) {
      if (token) {
        // The error will only be meaningful if a token was given
        dispatch(authenticationError(e.toString()));
      } else {
        dispatch(setAuthenticated(false));
      }
    }
  };
}

export function logout(): ThunkAction<Promise<void>, IStoreState, null, AuthAction> {
  return async dispatch => {
    Auth.unsetAuthToken();
    dispatch(setAuthenticated(false));
  };
}

export function tryToAutoAuthenticate(): ThunkAction<Promise<void>, IStoreState, null, AuthAction> {
  return async dispatch => {
    dispatch(checkingOIDCToken());
    dispatch(authenticate(undefined));
  };
}
