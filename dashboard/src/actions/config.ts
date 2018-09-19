import { Dispatch } from "redux";
import { ActionType, createActionDeprecated } from "typesafe-actions";

import Config, { IConfig } from "../shared/Config";

export const requestConfig = createActionDeprecated("REQUEST_CONFIG");
export const receiveConfig = createActionDeprecated("RECEIVE_CONFIG", (config: IConfig) => ({
  config,
  type: "RECEIVE_CONFIG",
}));

const allActions = [requestConfig, receiveConfig];
export type ConfigAction = ActionType<typeof allActions[number]>;

export function getConfig() {
  return async (dispatch: Dispatch) => {
    dispatch(requestConfig());
    const config = await Config.getConfig();
    return dispatch(receiveConfig(config));
  };
}
