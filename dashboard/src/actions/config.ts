import { ThunkAction } from "redux-thunk";
import { ActionType, createActionDeprecated } from "typesafe-actions";
import { IStoreState } from "../shared/types";

import Config, { IConfig } from "../shared/Config";

export const requestConfig = createActionDeprecated("REQUEST_CONFIG");
export const receiveConfig = createActionDeprecated("RECEIVE_CONFIG", (config: IConfig) => ({
  config,
  type: "RECEIVE_CONFIG",
}));

const allActions = [requestConfig, receiveConfig];
export type ConfigAction = ActionType<typeof allActions[number]>;

export function getConfig(): ThunkAction<Promise<void>, IStoreState, null, ConfigAction> {
  return async dispatch => {
    dispatch(requestConfig());
    const config = await Config.getConfig();
    dispatch(receiveConfig(config));
  };
}
