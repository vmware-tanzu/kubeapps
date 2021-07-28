import { ThunkAction } from "redux-thunk";
import { ActionType, deprecated } from "typesafe-actions";
import { IStoreState } from "../shared/types";

import Config, { IConfig, SupportedThemes } from "../shared/Config";

const { createAction } = deprecated;

export const requestConfig = createAction("REQUEST_CONFIG");
export const receiveConfig = createAction("RECEIVE_CONFIG", resolve => {
  return (config: IConfig) => resolve(config);
});
export const receiveTheme = createAction("RECEIVE_THEME", resolve => {
  return (theme: SupportedThemes) => resolve(theme);
});
export const errorConfig = createAction("ERROR_CONFIG", resolve => {
  return (err: Error) => resolve(err);
});

const allActions = [requestConfig, receiveConfig, receiveTheme, errorConfig];
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

// getTheme retrieves the current server configuration, calculates the proper theme
// and stores it in both the redux state and the user's localstorage
export function getTheme(): ThunkAction<Promise<void>, IStoreState, null, ConfigAction> {
  return async dispatch => {
    try {
      const config = await Config.getConfig();
      const theme = Config.getTheme(config);
      Config.setTheme(theme);
      dispatch(receiveTheme(theme));
    } catch (e) {
      dispatch(errorConfig(e));
    }
  };
}

// setUserTheme receives a theme and and stores it
// in both the redux state and the user's localstorage
export function setUserTheme(
  theme: SupportedThemes,
): ThunkAction<Promise<void>, IStoreState, null, ConfigAction> {
  return async dispatch => {
    Config.setUserTheme(theme);
    dispatch(receiveTheme(theme));
  };
}
