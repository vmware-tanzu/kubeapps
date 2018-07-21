import axios from "axios";

interface IConfig {
  namespace: string;
}

export default class Config {
  public static async getConfig() {
    if (Config.config) {
      return Config.config;
    }
    const url = `${Config.APIEndpoint}`;
    const { data } = await axios.get<IConfig>(url);
    Config.config = data;
    return Config.config;
  }

  private static config: IConfig;
  private static APIEndpoint: string = "/config.json";
}
