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
  theme: string;
}

export default class Config {
  public static async getConfig() {
    const url = Config.APIEndpoint;
    const { data } = await axios.get<IConfig>(url);
    return data;
  }

  public static getTheme(config: IConfig): SupportedThemes {
    // Define a ballback theme in case of errors
    const fallbackTheme = SupportedThemes.light;

    // Retrieve the system theme preference (configurable via Values.dashboard.defaultTheme)
    const systemTheme = config.theme != null ? SupportedThemes[config.theme] : undefined;

    // Retrieve the user theme preference
    const userTheme =
      localStorage.getItem("theme") != null
        ? SupportedThemes[localStorage.getItem("theme") as string]
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

  public static setTheme(theme: SupportedThemes) {
    document.body.setAttribute("cds-theme", theme);
    localStorage.setItem("theme", theme);
  }

  private static APIEndpoint = "config.json";
}
