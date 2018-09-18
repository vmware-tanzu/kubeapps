import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import { IStoreState } from "shared/types";
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
  return async (dispatch: ThunkDispatch<IStoreState, null, Action>) => {
    dispatch(requestConfig);
    const config = await Config.getConfig();
    return dispatch(receiveConfig(config));
  };
}
