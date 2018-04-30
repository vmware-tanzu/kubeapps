import { axios } from "./Auth";

import { IKubelessConfigMap } from "./types";

export default class Config {
  public static async get() {
    const { data: config } = await axios.get<IKubelessConfigMap>(Config.SelfLink);
    return config;
  }

  public static async getRuntimes() {
    const config = await this.get();
    return JSON.parse(config.data["runtime-images"]);
  }

  private static Name: string = "kubeless-config";
  private static Namespace: string = "kubeless";
  private static APIBase: string = "/api/kube";
  private static APIEndpoint: string = `${Config.APIBase}/api/v1`;
  private static SelfLink: string = `${Config.APIEndpoint}/namespaces/${
    Config.Namespace
  }/configmaps/${Config.Name}`;
}
