// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import Axios, { AxiosError, AxiosInstance, AxiosRequestConfig, AxiosResponse } from "axios";
import { Action, Store } from "redux";
import { ThunkDispatch } from "redux-thunk";
import actions from "../actions";
import { Auth } from "./Auth";
import {
  ConflictNetworkError,
  ForbiddenNetworkError,
  InternalServerNetworkError,
  IRBACRole,
  IStoreState,
  NotFoundNetworkError,
  UnauthorizedNetworkError,
  UnprocessableEntityError,
} from "./types";

export function addAuthHeaders(axiosInstance: AxiosInstance) {
  axiosInstance.interceptors.request.use((config: AxiosRequestConfig) => {
    const authToken = Auth.getAuthToken();
    if (authToken && config?.headers) {
      config.headers.Authorization = `Bearer ${authToken}`;
    }
    return config;
  });
}

export function addErrorHandling(axiosInstance: AxiosInstance, store: Store<IStoreState>) {
  axiosInstance.interceptors.response.use(
    response => response,
    e => {
      const dispatch = store.dispatch as ThunkDispatch<IStoreState, null, Action>;
      const err: AxiosError<any, any> = e;
      if (
        err.code === undefined &&
        err.message === "Network Error" &&
        !err.response &&
        Auth.usingOIDCToken()
      ) {
        // The OIDC token is no longer valid, logout
        dispatch(actions.auth.expireSession());
      }
      let message = err.message;
      if (err.response) {
        if (err.response.data.message) {
          message = err.response.data.message;
        }
        if (typeof err.response.data === "string") {
          message = err.response.data;
        }
      }

      const dispatchErrorAndLogout = (m: string) => {
        // Global action dispatch to log the user out
        dispatch(actions.auth.authenticationError(m));
        // Expire the session if we are using OIDC tokens and
        // logout either way.
        dispatch(actions.auth.expireSession());
      };
      const response = err.response as AxiosResponse<any>;
      switch (response && response.status) {
        case 401:
          dispatchErrorAndLogout(message);
          return Promise.reject(new UnauthorizedNetworkError(message));
        case 403:
          // Subcase 1:
          //   if usingOIDCToken: a 403 directly from the auth proxy
          //   always requires reauthentication.
          // Subcase 2:
          //   if !usingOIDCToken, an anonymous response is sent when
          //   a serviceaccount token expires (eg., delete secret xxx)
          //   or the token is managed externally (eg., vsphere kubectl login)
          //   In this case, force reauthentication
          if (Auth.is403FromAuthProxy(response) || Auth.isAnonymous(response)) {
            dispatchErrorAndLogout(message);
          }
          // Subcase 3:
          //   The most likely case is just a 403 due to a lack of
          //   permissions in a certain namespace.
          //   In this case, just return the error message back to the user
          try {
            const jsonMessage = JSON.parse(message) as IRBACRole[];
            return Promise.reject(
              new ForbiddenNetworkError(
                `Forbidden error, missing permissions: ${jsonMessage
                  .map(forbiddenAction => {
                    const { apiGroup, resource, namespace, clusterWide, verbs } = forbiddenAction;
                    return `apiGroup: "${apiGroup}", resource: "${resource}", action: "${verbs.join(
                      ", ",
                    )}", ${clusterWide ? "in all namespaces" : `namespace: ${namespace}`}`;
                  })
                  .join("; ")}`,
              ),
            );
          } catch (e: any) {
            // Subcase 4:
            //   A non-parseable 403 error.
            //   Do not require reauthentication and display error (ie. edge cases of proxy auth)
          }
          return Promise.reject(new ForbiddenNetworkError(message));
        case 404:
          return Promise.reject(new NotFoundNetworkError(message));
        case 409:
          return Promise.reject(new ConflictNetworkError(message));
        case 422:
          return Promise.reject(new UnprocessableEntityError(message));
        case 500:
          return Promise.reject(new InternalServerNetworkError(message));
        default:
          return Promise.reject(new Error(message));
      }
    },
  );
}

// Error handling is added with an interceptor in index.tsx
export const axios = Axios.create();
// Authorization headers and error handling are added with an interceptor in index.tsx
export const axiosWithAuth = Axios.create();
