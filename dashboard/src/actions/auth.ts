import { ThunkAction } from "redux-thunk";
import { ActionType, createAction } from "typesafe-actions";

import { Auth } from "../shared/Auth";
import { IStoreState } from "../shared/types";

export const setAuthenticated = createAction("SET_AUTHENTICATED", resolve => {
  return (authenticated: boolean, oidc: boolean) => resolve({ authenticated, oidc });
});

export const authenticating = createAction("AUTHENTICATING");

export const authenticationError = createAction("AUTHENTICATION_ERROR", resolve => {
  return (errorMsg: string) => resolve(errorMsg);
});

export const setSessionExpired = createAction("SET_AUTHENTICATION_SESSION_EXPIRED", resolve => {
  return (sessionExpired: boolean) => resolve({ sessionExpired });
});

const allActions = [setAuthenticated, authenticating, authenticationError, setSessionExpired];

export type AuthAction = ActionType<typeof allActions[number]>;

export function authenticate(
  token: string,
  oidc: boolean,
): ThunkAction<Promise<void>, IStoreState, null, AuthAction> {
  return async dispatch => {
    dispatch(authenticating());
    try {
      await Auth.validateToken(token);
      Auth.setAuthToken(token, oidc);
      dispatch(setAuthenticated(true, oidc));
      if (oidc) {
        dispatch(setSessionExpired(false));
      }
    } catch (e) {
      dispatch(authenticationError(e.toString()));
    }
  };
}

export function logout(): ThunkAction<Promise<void>, IStoreState, null, AuthAction> {
  return async dispatch => {
    Auth.unsetAuthToken();
    dispatch(setAuthenticated(false, false));
  };
}

export function expireSession(): ThunkAction<Promise<void>, IStoreState, null, AuthAction> {
  return async dispatch => {
    if (Auth.usingOIDCToken()) {
      dispatch(setSessionExpired(true));
    }
    return dispatch(logout());
  };
}

export function tryToAuthenticateWithOIDC(): ThunkAction<
  Promise<void>,
  IStoreState,
  null,
  AuthAction
> {
  return async dispatch => {
    dispatch(authenticating());
    const token = await Auth.fetchOIDCToken();
    if (token) {
      dispatch(authenticate(token, true));
    } else {
      dispatch(setAuthenticated(false, false));
    }
  };
}
