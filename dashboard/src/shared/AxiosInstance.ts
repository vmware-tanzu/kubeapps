import Axios, { AxiosError, AxiosInstance, AxiosRequestConfig } from "axios";

import { Action, Store } from "redux";
import { ThunkDispatch } from "redux-thunk";
import actions from "../actions";
import { Auth } from "./Auth";
import {
  ConflictError,
  ForbiddenError,
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
      if (err.response && err.response.data.message) {
        message = err.response.data.message;
      }
      switch (err.response && err.response.status) {
        case 401:
          // Global action dispatch to log the user out
          // Expire the session if we are using OIDC tokens
          dispatch(actions.auth.authenticationError(message));
          dispatch(actions.auth.expireSession());
          return Promise.reject(new UnauthorizedError(message));
        case 403:
          return Promise.reject(new ForbiddenError(message));
        case 404:
          return Promise.reject(new NotFoundError(message));
        case 409:
          return Promise.reject(new ConflictError(message));
        case 422:
          return Promise.reject(new UnprocessableEntity(message));
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
