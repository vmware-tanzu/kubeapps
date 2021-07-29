import { getType } from "typesafe-actions";
import actions from "../actions";
import { ConfigAction } from "../actions/config";
import { IConfig, SupportedThemes } from "../shared/Config";

export interface IConfigState extends IConfig {
  loaded: boolean;
}

export const initialState: IConfigState = {
  loaded: false,
  kubeappsCluster: "",
  kubeappsNamespace: "",
  appVersion: "",
  authProxyEnabled: false,
  oauthLoginURI: "",
  oauthLogoutURI: "",
  authProxySkipLoginPage: false,
  clusters: [],
  theme: SupportedThemes.light,
};

const configReducer = (state: IConfigState = initialState, action: ConfigAction): IConfigState => {
  switch (action.type) {
    case getType(actions.config.requestConfig):
      return initialState;
    case getType(actions.config.receiveConfig):
      return {
        ...state,
        loaded: true,
        ...action.payload,
      };
    case getType(actions.config.receiveTheme):
      return {
        ...state,
        theme: action.payload,
      };
    case getType(actions.config.errorConfig):
      return {
        ...state,
        error: action.payload,
      };
    default:
  }
  return state;
};

export default configReducer;
