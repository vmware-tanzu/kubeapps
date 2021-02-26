import { ThunkAction } from "redux-thunk";
import { ActionType, deprecated } from "typesafe-actions";
import { IStoreState } from "../shared/types";

import Config, { IConfig } from "../shared/Config";

const { createAction } = deprecated;

export const requestConfig = createAction("REQUEST_CONFIG");
export const receiveConfig = createAction("RECEIVE_CONFIG", resolve => {
  return (config: IConfig) => resolve(config);
});
export const errorConfig = createAction("ERROR_CONFIG", resolve => {
  return (err: Error) => resolve(err);
});

const allActions = [requestConfig, receiveConfig, errorConfig];
export type ConfigAction = ActionType<typeof allActions[number]>;

export function getConfig(): ThunkAction<Promise<void>, IStoreState, null, ConfigAction> {
  return async dispatch => {
    dispatch(requestConfig());
    try {
      const config = await Config.getConfig();
      dispatch(receiveConfig(config));
    } catch (e) {
      dispatch(errorConfig(e));
    }
  };
}
