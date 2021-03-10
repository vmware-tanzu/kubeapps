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
export const setThemeState = createAction("SET_THEME", resolve => {
  return (theme: SupportedThemes) => resolve(theme);
});
export const errorConfig = createAction("ERROR_CONFIG", resolve => {
  return (err: Error) => resolve(err);
});

const allActions = [requestConfig, receiveConfig, receiveTheme, setThemeState, errorConfig];
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

export function getTheme(): ThunkAction<Promise<void>, IStoreState, null, ConfigAction> {
  return async dispatch => {
    const theme = Config.getTheme();
    dispatch(receiveTheme(theme));
  };
}

export function setTheme(
  theme: SupportedThemes,
): ThunkAction<Promise<void>, IStoreState, null, ConfigAction> {
  return async dispatch => {
    Config.setTheme(theme);
    dispatch(setThemeState(theme));
  };
}
