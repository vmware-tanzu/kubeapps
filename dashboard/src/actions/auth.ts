import { Dispatch } from "redux";
import { createAction, getReturnOfExpression } from "typesafe-actions";

import { IStoreState } from "../shared/types";

export const setAuthenticated = createAction("SET_AUTHENTICATED");

const allActions = [setAuthenticated].map(getReturnOfExpression);
export type AuthAction = typeof allActions[number];

export function authenticate(token: string) {
  return async (dispatch: Dispatch<IStoreState>) => {
    localStorage.setItem("kubeapps_auth_token", token);
    return dispatch(setAuthenticated());
  };
}
