import axios from "axios";

export enum SupportedThemes {
  dark = "dark",
  light = "light",
}

// IConfig is the configuration for Kubeapps
export interface IConfig {
  kubeappsCluster: string;
  kubeappsNamespace: string;
  appVersion: string;
  authProxyEnabled: boolean;
  oauthLoginURI: string;
  oauthLogoutURI: string;
  authProxySkipLoginPage: boolean;
  error?: Error;
  clusters: string[];
  theme: SupportedThemes;
}

export default class Config {
  public static async getConfig() {
    const url = Config.APIEndpoint;
    const { data } = await axios.get<IConfig>(url);
    return data;
  }

  public static getTheme() {
    let theme = localStorage.getItem("theme");
    if (!theme) {
      theme =
        window.matchMedia && window.matchMedia("(prefers-color-scheme: dark)").matches
          ? SupportedThemes.dark
          : SupportedThemes.light;
    }
    return (theme as SupportedThemes) || SupportedThemes.light;
  }

  public static setTheme(theme: SupportedThemes) {
    document.body.setAttribute("cds-theme", theme);
    localStorage.setItem("theme", theme);
  }

  private static APIEndpoint = "config.json";
}
