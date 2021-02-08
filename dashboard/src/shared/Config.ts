import axios from "axios";

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
}

export default class Config {
  public static async getConfig() {
    const url = Config.APIEndpoint;
    const { data } = await axios.get<IConfig>(url);
    return data;
  }

  private static APIEndpoint: string = "config.json";
}
