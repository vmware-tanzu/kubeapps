import { AxiosError, AxiosInstance, AxiosRequestConfig } from "axios";
import { Store } from "redux";
import actions from "../actions";
import { Auth } from "./Auth";
import {
  AppConflict,
  ForbiddenError,
  IStoreState,
  NotFoundError,
  UnauthorizedError,
  UnprocessableEntity,
} from "./types";

// authenticatedAxiosInstance returns an axios instance with an interceptor
// configured to set the current auth token and handle errors.
export function createAxiosInterceptors(axios: AxiosInstance, store: Store<IStoreState>) {
  axios.interceptors.request.use((config: AxiosRequestConfig) => {
    const authToken = Auth.getAuthToken();
    if (authToken) {
      config.headers.Authorization = `Bearer ${authToken}`;
    }
    return config;
  });
  axios.interceptors.response.use(
    response => response,
    e => {
      const err: AxiosError = e;
      let message = err.message;
      if (err.response && err.response.data.message) {
        message = err.response.data.message;
      }
      switch (err.response && err.response.status) {
        case 401:
          // Global action dispatch to log the user out
          if (err.response) {
            store.dispatch(actions.auth.authenticationError(message));
            store.dispatch(actions.auth.logout());
          }
          return Promise.reject(new UnauthorizedError(message));
        case 403:
          return Promise.reject(new ForbiddenError(message));
        case 404:
          return Promise.reject(new NotFoundError(message));
        case 409:
          return Promise.reject(new AppConflict(message));
        case 422:
          return Promise.reject(new UnprocessableEntity(message));
        default:
          return Promise.reject(e);
      }
    },
  );
}
