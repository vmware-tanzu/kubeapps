import { axiosWithAuth } from "./AxiosInstance";
import { APIBase } from "./Kube";

import { IK8sList, IResource } from "./types";

export default class Namespace {
  public static async list() {
    const { data } = await axiosWithAuth.get<IK8sList<IResource, {}>>(`${Namespace.APIEndpoint}`);
    return data;
  }

  public static async create(name: string) {
    const { data } = await axiosWithAuth.post<IResource>(Namespace.APIEndpoint, {
      apiVersion: "v1",
      kind: "Namespace",
      metadata: {
        name,
      },
    });
    return data;
  }

  private static APIBase: string = APIBase;
  private static APIEndpoint: string = `${Namespace.APIBase}/api/v1/namespaces/`;
}

// Set of namespaces used accross the applications as default and "all ns" placeholders
export const definedNamespaces = {
  default: "default",
  all: "_all",
};
