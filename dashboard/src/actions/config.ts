// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { ThunkAction } from "redux-thunk";
import Config, { IConfig, SupportedThemes } from "shared/Config";
import { IStoreState } from "shared/types";
import { ActionType, deprecated } from "typesafe-actions";

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
      const configuredPlugins = await Config.getConfiguredPlugins();
      dispatch(receiveConfig({ ...config, configuredPlugins }));
    } catch (e: any) {
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
    } catch (e: any) {
      dispatch(errorConfig(e));
    }
  };
}

// setUserTheme receives a theme and stores it
// in both the redux state and the user's localstorage
export function setUserTheme(
  theme: SupportedThemes,
): ThunkAction<Promise<void>, IStoreState, null, ConfigAction> {
  return async dispatch => {
    Config.setUserTheme(theme);
    dispatch(receiveTheme(theme));
  };
}
