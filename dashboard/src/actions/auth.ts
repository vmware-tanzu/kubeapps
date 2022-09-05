// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { ThunkAction } from "redux-thunk";
import { Auth } from "shared/Auth";
import * as Namespace from "shared/Namespace";
import { IStoreState, UnauthorizedNetworkError } from "shared/types";
import { ActionType, deprecated } from "typesafe-actions";
import { clearClusters, NamespaceAction } from "./namespace";

const { createAction } = deprecated;

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
  cluster: string,
  token: string,
  oidc: boolean,
): ThunkAction<Promise<void>, IStoreState, null, AuthAction> {
  return async dispatch => {
    dispatch(authenticating());
    try {
      if (!oidc) {
        await Auth.validateToken(cluster, token);
      }
      Auth.setAuthToken(token, oidc);
      dispatch(setAuthenticated(true, oidc));
      if (oidc) {
        dispatch(setSessionExpired(false));
      }
    } catch (e: any) {
      dispatch(authenticationError(e.toString()));
    }
  };
}

export function logout(): ThunkAction<
  Promise<void>,
  IStoreState,
  null,
  AuthAction | NamespaceAction
> {
  return async (dispatch, getState) => {
    // We can't do anything before calling unsetAuthCookie as otherwise the
    // state changes and the redirect to the logout URI is lost.
    if (Auth.usingOIDCToken()) {
      const { config } = getState();
      Auth.unsetAuthCookie(config);
    } else {
      Auth.unsetAuthToken();
      dispatch(setAuthenticated(false, false));
      dispatch(clearClusters());
    }
    Namespace.unsetStoredNamespace();
  };
}

export function logoutByAuthenticationError(): ThunkAction<
  Promise<void>,
  IStoreState,
  null,
  AuthAction | NamespaceAction
> {
  return async dispatch => {
    dispatch(logout());
    dispatch(authenticationError("Unauthorized"));
    dispatch(expireSession());
  };
}

export function handleErrorAction(error: any, action?: ActionType<any>) {
  if (error.constructor === UnauthorizedNetworkError) {
    return logoutByAuthenticationError();
  } else if (action) {
    return action;
  }
}

export function expireSession(): ThunkAction<Promise<void>, IStoreState, null, AuthAction> {
  return async dispatch => {
    if (Auth.usingOIDCToken()) {
      dispatch(setSessionExpired(true));
    }
    return dispatch(logout());
  };
}

export function checkCookieAuthentication(
  cluster: string,
): ThunkAction<Promise<boolean>, IStoreState, null, AuthAction> {
  return async dispatch => {
    // The call to authenticate below will also dispatch authenticating,
    // but we dispatch it early so that the login screen is shown as
    // loading while we query isAuthenticatedWithCookie().
    dispatch(authenticating());
    const isAuthed = await Auth.isAuthenticatedWithCookie(cluster);
    if (isAuthed) {
      await dispatch(authenticate(cluster, "", true));
    } else {
      dispatch(setAuthenticated(false, false));
    }
    return isAuthed;
  };
}
