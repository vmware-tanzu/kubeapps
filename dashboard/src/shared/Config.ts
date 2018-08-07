import axios from "axios";

// IConfig is the configuration for Kubeapps
export interface IConfig {
  namespace: string;
}

export default class Config {
  public static async getConfig() {
    const url = Config.APIEndpoint;
    const { data } = await axios.get<IConfig>(url);

    // Development environment config overrides
    // TODO(miguel) Rename env variable to TELEPRESENCE_CONTAINER_NAMESPACE
    // and remove package.json yarn run mapping once create-react-app is ejected
    if (process.env.REACT_APP_KUBEAPPS_NS) {
      data.namespace = process.env.REACT_APP_KUBEAPPS_NS;
    }

    return data;
  }

  private static APIEndpoint: string = "/config.json";
}
