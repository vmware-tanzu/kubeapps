// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { IConfig, SupportedThemes } from "shared/Config";
import { getType } from "typesafe-actions";
import actions from "../actions";
import { ConfigAction } from "../actions/config";

export interface IConfigState extends IConfig {
  loaded: boolean;
}

export const initialState: IConfigState = {
  loaded: false,
  kubeappsCluster: "",
  kubeappsNamespace: "",
  helmGlobalNamespace: "",
  carvelGlobalNamespace: "",
  appVersion: "",
  authProxyEnabled: false,
  oauthLoginURI: "",
  oauthLogoutURI: "",
  authProxySkipLoginPage: false,
  clusters: [],
  featureFlags: { operators: false },
  theme: SupportedThemes.light,
  remoteComponentsUrl: "",
  customAppViews: [],
  skipAvailablePackageDetails: false,
  createNamespaceLabels: {},
  configuredPlugins: [],
};

const configReducer = (state: IConfigState = initialState, action: ConfigAction): IConfigState => {
  switch (action.type) {
    case getType(actions.config.requestConfig):
      return initialState;
    case getType(actions.config.receiveConfig):
      return {
        ...state,
        loaded: true,
        ...action.payload,
      };
    case getType(actions.config.receiveTheme):
      return {
        ...state,
        theme: action.payload,
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
