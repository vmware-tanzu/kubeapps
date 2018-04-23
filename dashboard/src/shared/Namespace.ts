import { axios } from "./Auth";

import { IK8sList, IResource } from "./types";

export default class Namespace {
  public static async list() {
    const { data } = await axios.get<IK8sList<IResource, {}>>(`${Namespace.APIEndpoint}`);
    return data;
  }

  private static APIBase: string = "/api/kube";
  private static APIEndpoint: string = `${Namespace.APIBase}/api/v1/namespaces`;
}
