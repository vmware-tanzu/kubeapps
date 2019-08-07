import { axiosWithAuth } from "./AxiosInstance";
import { APIBase } from "./Kube";
import { definedNamespaces } from "./Namespace";
import { IOwnerReference, ISecret } from "./types";

export default class Secret {
  public static async create(
    name: string,
    secrets: { [s: string]: string },
    owner: IOwnerReference | undefined,
    namespace: string = definedNamespaces.default,
  ) {
    const url = Secret.getLink(namespace);
    const { data } = await axiosWithAuth.post<ISecret>(url, {
      apiVersion: "v1",
      data: secrets,
      kind: "Secret",
      metadata: {
        name,
        ownerReferences: [owner],
      },
      type: "Opaque",
    });
    return data;
  }

  public static async delete(name: string, namespace: string = definedNamespaces.default) {
    const url = this.getLink(namespace, name);
    return axiosWithAuth.delete(url);
  }

  public static async get(name: string, namespace: string = definedNamespaces.default) {
    const url = this.getLink(namespace, name);
    const { data } = await axiosWithAuth.get<ISecret>(url);
    return data;
  }

  public static async list(namespace: string = definedNamespaces.default) {
    const url = Secret.getLink(namespace);
    const { data } = await axiosWithAuth.get<ISecret>(url);
    return data;
  }

  private static getLink(namespace: string, name?: string): string {
    return `${APIBase}/api/v1/namespaces/${namespace}/secrets${name ? `/${name}` : ""}`;
  }
}
