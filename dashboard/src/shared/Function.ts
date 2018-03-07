import axios from "axios";

import { IFunction, IFunctionList } from "./types";

export default class Function {
  public static async list() {
    const { data } = await axios.get<IFunctionList>(Function.APIEndpoint);
    return data;
  }

  public static async get(name: string) {
    const { data } = await axios.get(Function.getSelfLink(name));
    return data;
  }

  public static async update(name: string, newApp: IFunction) {
    const { data } = await axios.put(Function.getSelfLink(name), newApp);
    return data;
  }

  public static async delete(name: string) {
    const { data } = await axios.delete(Function.getSelfLink(name));
    return data;
  }

  //   public static async create(name: string, url: string) {
  //   }

  private static APIEndpoint: string = "/api/kube/apis/kubeless.io/v1beta1/functions";
  private static getSelfLink(name: string): string {
    return `${Function.APIEndpoint}/${name}`;
  }
}
