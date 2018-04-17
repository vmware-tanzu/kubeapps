import { Dispatch } from "redux";
import { createAction, getReturnOfExpression } from "typesafe-actions";

import { Auth } from "../shared/Auth";
import { IStoreState } from "../shared/types";

export const setAuthenticated = createAction("SET_AUTHENTICATED", (authenticated: boolean) => ({
  authenticated,
  type: "SET_AUTHENTICATED",
}));

const allActions = [setAuthenticated].map(getReturnOfExpression);
export type AuthAction = typeof allActions[number];

export function authenticate(token: string) {
  return async (dispatch: Dispatch<IStoreState>) => {
    await Auth.validateToken(token);
    Auth.setAuthToken(token);
    return dispatch(setAuthenticated(true));
  };
}

export function logout() {
  return async (dispatch: Dispatch<IStoreState>) => {
    Auth.unsetAuthToken();
    return dispatch(setAuthenticated(false));
  };
}
