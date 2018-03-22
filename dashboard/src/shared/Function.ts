import axios from "axios";

import { IFunction, IFunctionList, IResource, IStatus } from "./types";

export default class Function {
  public static async list() {
    const { data } = await axios.get<IFunctionList>(`${Function.APIEndpoint}/functions`);
    return data;
  }

  public static async get(name: string, namespace: string) {
    const { data } = await axios.get(Function.getSelfLink(name, namespace));
    return data;
  }

  public static async getPodName(fn: IFunction) {
    const { data: { items } } = await axios.get<{ items: IResource[] }>(
      `${Function.APIBase}/api/v1/namespaces/${fn.metadata.namespace}/pods?labelSelector=function=${
        fn.metadata.name
      }`,
    );
    // find the first pod that isn't terminating
    // kubectl uses deletionTimestamp to determine Terminating status, pod phase does not report this
    const pod = items.find(i => !i.metadata.deletionTimestamp);
    if (pod) {
      return pod.metadata.name;
    }
    return;
  }

  public static async create(name: string, namespace: string, spec: IFunction["spec"]) {
    try {
      const { data } = await axios.post<IFunction>(
        `${Function.APIEndpoint}/namespaces/${namespace}/functions`,
        {
          apiVersion: "kubeless.io/v1beta1",
          kind: "Function",
          metadata: {
            name,
          },
          spec,
        },
      );
      return data;
    } catch (err) {
      throw new Error((err.response.data as IStatus).message);
    }
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

  private static APIBase: string = "/api/kube";
  private static APIEndpoint: string = `${Function.APIBase}/apis/kubeless.io/v1beta1`;
  private static getSelfLink(name: string, namespace: string): string {
    return `${Function.APIEndpoint}/namespaces/${namespace}/functions/${name}`;
  }
}
