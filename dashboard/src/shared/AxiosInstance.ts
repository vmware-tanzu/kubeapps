import { AxiosError, AxiosInstance, AxiosRequestConfig } from "axios";
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

// createAxiosInterceptors will configure a set of interceptors to a provided axios instance,
// relying also on an external redux store for action dispatching
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
          const dispatch = store.dispatch as ThunkDispatch<IStoreState, null, Action>;
          dispatch(actions.auth.authenticationError(message));
          dispatch(actions.auth.logout());
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
