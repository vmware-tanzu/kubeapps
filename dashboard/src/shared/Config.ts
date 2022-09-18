// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import axios from "axios";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import * as url from "shared/url";
import { PackageRepositoriesService } from "./PackageRepositoriesService";

export enum SupportedThemes {
  dark = "dark",
  light = "light",
}

export interface ICustomAppViewIdentifier {
  name: string;
  plugin: string;
  repository: string;
}

// IConfig is the configuration for Kubeapps
export interface IConfig {
  kubeappsCluster: string;
  kubeappsNamespace: string;
  helmGlobalNamespace: string;
  // TODO(castelblanque) Global namespaces should be well organized by plugin, or come from plugins API
  carvelGlobalNamespace: string;
  appVersion: string;
  authProxyEnabled: boolean;
  oauthLoginURI: string;
  oauthLogoutURI: string;
  authProxySkipLoginPage: boolean;
  error?: Error;
  clusters: string[];
  featureFlags: IFeatureFlags;
  theme: string;
  remoteComponentsUrl: string;
  customAppViews: ICustomAppViewIdentifier[];
  skipAvailablePackageDetails: boolean;
  createNamespaceLabels: { [key: string]: string };
  configuredPlugins: Plugin[];
}

export interface IFeatureFlags {
  operators: boolean;
}

export default class Config {
  public static async getConfig() {
    const { data } = await axios.get<IConfig>(url.api.config);
    return data;
  }

  public static async getConfiguredPlugins() {
    const { plugins } = await PackageRepositoriesService.getConfiguredPlugins();
    return plugins;
  }

  // getTheme retrieves the different theme preferences and calculates which one is chosen
  public static getTheme(config: IConfig): SupportedThemes {
    // Define a ballback theme in case of errors
    const fallbackTheme = SupportedThemes.light;

    // Retrieve the system theme preference (configurable via Values.dashboard.defaultTheme)
    const systemTheme = config.theme != null ? SupportedThemes[config.theme] : undefined;

    // Retrieve the user theme preference
    const userTheme =
      localStorage.getItem("user-theme") != null
        ? SupportedThemes[localStorage.getItem("user-theme") as string]
        : undefined;

    // Retrieve the browser theme preference
    const browserTheme =
      window.matchMedia && window.matchMedia("(prefers-color-scheme: dark)").matches
        ? SupportedThemes.dark
        : SupportedThemes.light;

    // calculates the chose theme based upon this prelation order: user>system>browser>fallback
    const chosenTheme = userTheme ?? systemTheme ?? browserTheme ?? fallbackTheme;

    return chosenTheme;
  }

  // setTheme performs a hot change of the current theme modifying the DOM
  // it's a separate function for testing
  public static setTheme(theme: SupportedThemes) {
    document.body.setAttribute("cds-theme", theme);
  }

  // setUserTheme changes the current theme and also stores the user's preference in the localStorage
  // it's a separate function for testing
  public static setUserTheme(theme: SupportedThemes) {
    Config.setTheme(theme);
    localStorage.setItem("user-theme", theme);
  }
}
