import axios from "axios";

import { IFunction, IFunctionList } from "./types";

export default class Function {
  public static async list() {
    const { data } = await axios.get<IFunctionList>(`${Function.APIEndpoint}/functions`);
    return data;
  }

  public static async get(name: string, namespace: string) {
    const { data } = await axios.get(Function.getSelfLink(name, namespace));
    return data;
  }

  public static async update(name: string, namespace: string, newFn: IFunction) {
    const { data } = await axios.put(Function.getSelfLink(name, namespace), newFn);
    return data;
  }

  public static async delete(name: string, namespace: string) {
    const { data } = await axios.delete(Function.getSelfLink(name, namespace));
    return data;
  }

  //   public static async create(name: string, url: string) {
  //   }

  private static APIEndpoint: string = "/api/kube/apis/kubeless.io/v1beta1";
  private static getSelfLink(name: string, namespace: string): string {
    return `${Function.APIEndpoint}/namespaces/${namespace}/functions/${name}`;
  }
}
