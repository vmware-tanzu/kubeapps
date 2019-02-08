import axios from "axios";

// IConfig is the configuration for Kubeapps
export interface IConfig {
  namespace: string;
  appVersion: string;
  error?: Error;
}

export default class Config {
  public static async getConfig() {
    const url = Config.APIEndpoint;
    const { data } = await axios.get<IConfig>(url);

    // Development environment config overrides
    // TODO(miguel) Rename env variable to KUBEAPPS_NAMESPACE once/if we eject create-react-app
    // Currently we are using REACT_APP_* because it's the only way to inject env variables in a sealed setup.
    // Please note that this env variable gets mapped in the run command in the package.json file
    if (process.env.NODE_ENV !== "production" && process.env.REACT_APP_KUBEAPPS_NS) {
      data.namespace = process.env.REACT_APP_KUBEAPPS_NS;
      data.appVersion = "DEVEL";
    }

    return data;
  }

  private static APIEndpoint: string = "config.json";
}
