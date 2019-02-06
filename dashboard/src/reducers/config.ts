import { getType } from "typesafe-actions";

import actions from "../actions";
import { ConfigAction } from "../actions/config";
import { IConfig } from "../shared/Config";

export interface IConfigState extends IConfig {
  loaded: boolean;
}

const initialState: IConfigState = {
  loaded: false,
  namespace: "",
  appVersion: "",
};

const configReducer = (state: IConfigState = initialState, action: ConfigAction): IConfigState => {
  switch (action.type) {
    case getType(actions.config.requestConfig):
      return initialState;
    case getType(actions.config.receiveConfig):
      return {
        loaded: true,
        ...action.payload,
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
