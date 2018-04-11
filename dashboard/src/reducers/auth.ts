import { getType } from "typesafe-actions";

import actions from "../actions";
import { AuthAction } from "../actions/auth";

export interface IAuthState {
  authenticated: boolean;
}

const initialState: IAuthState = {
  authenticated: !(localStorage.getItem("kubeapps_auth_token") === null),
};

const authReducer = (state: IAuthState = initialState, action: AuthAction): IAuthState => {
  switch (action.type) {
    case getType(actions.auth.setAuthenticated):
      return { authenticated: true };
    default:
  }
  return state;
};

export default authReducer;
