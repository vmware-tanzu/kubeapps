import axios from "axios";

import { IKubelessConfigMap } from "./types";

export default class Function {
  public static async get() {
    const config = await axios.get<IKubelessConfigMap>(Function.SelfLink);
    return config;
  }

  public static async getRuntimes() {
    const config = await this.get();
    return JSON.parse(config.data.data["runtime-images"]);
  }

  private static Name: string = "kubeless-config";
  private static Namespace: string = "kubeless";
  private static APIBase: string = "/api/kube";
  private static APIEndpoint: string = `${Function.APIBase}/api/v1`;
  private static SelfLink: string = `${Function.APIEndpoint}/namespaces/${
    Function.Namespace
  }/configmaps/${Function.Name}`;
}
