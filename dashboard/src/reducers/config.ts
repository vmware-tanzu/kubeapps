import { getType } from "typesafe-actions";

import actions from "../actions";
import { ConfigAction } from "../actions/config";
import { IConfig } from "../shared/Config";

const initialState: IConfig = {
  namespace: "default",
};

const configReducer = (state: IConfig = initialState, action: ConfigAction): IConfig => {
  switch (action.type) {
    case getType(actions.config.receiveConfig):
      return action.config;
    default:
  }
  return state;
};

export default configReducer;
