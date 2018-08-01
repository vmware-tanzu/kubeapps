import { Dispatch } from "redux";
import { createAction, getReturnOfExpression } from "typesafe-actions";

import Config, { IConfig } from "../shared/Config";
import { IStoreState } from "../shared/types";

export const requestConfig = createAction("REQUEST_CONFIG");
export const receiveConfig = createAction("RECEIVE_CONFIG", (config: IConfig) => ({
  config,
  type: "RECEIVE_CONFIG",
}));

const allActions = [requestConfig, receiveConfig].map(getReturnOfExpression);
export type ConfigAction = typeof allActions[number];

export function getConfig() {
  return async (dispatch: Dispatch<IStoreState>) => {
    const config = await Config.getConfig();
    return dispatch(receiveConfig(config));
  };
}
