import Axios, { AxiosError, AxiosInstance, AxiosRequestConfig, AxiosResponse } from "axios";

import { Action, Store } from "redux";
import { ThunkDispatch } from "redux-thunk";
import actions from "../actions";
import { Auth } from "./Auth";
import {
  ConflictError,
  ForbiddenError,
  InternalServerError,
  IStoreState,
  NotFoundError,
  UnauthorizedError,
  UnprocessableEntity,
} from "./types";

export function addAuthHeaders(axiosInstance: AxiosInstance) {
  axiosInstance.interceptors.request.use((config: AxiosRequestConfig) => {
    const authToken = Auth.getAuthToken();
    if (authToken) {
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
      const err: AxiosError = e;
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
      const response = err.response as AxiosResponse;
      switch (response && response.status) {
        case 401:
          dispatchErrorAndLogout(message);
          return Promise.reject(new UnauthorizedError(message));
        case 403:
          // A 403 directly from the auth proxy requires reauthentication.
          if (Auth.usingOIDCToken() && Auth.is403FromAuthProxy(response)) {
            dispatchErrorAndLogout(message);
          }
          return Promise.reject(new ForbiddenError(message));
        case 404:
          return Promise.reject(new NotFoundError(message));
        case 409:
          return Promise.reject(new ConflictError(message));
        case 422:
          return Promise.reject(new UnprocessableEntity(message));
        case 500:
          return Promise.reject(new InternalServerError(message));
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
