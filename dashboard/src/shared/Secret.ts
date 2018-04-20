import axios from "axios";

import { IOwnerReference, IResource, ISecret, IStatus } from "./types";

export default class Secret {
  public static async create(
    name: string,
    secrets: { [s: string]: string },
    owner: any | undefined,
    namespace: string = "default",
  ) {
    const url = Secret.getLink(namespace);
    try {
      const { data } = await axios.post<ISecret>(url, {
        apiVersion: "v1",
        data: secrets,
        kind: "Secret",
        metadata: {
          name,
          ownerReferences: [Secret.getOwnerReference(owner)],
        },
        type: "Opaque",
      });
      return data;
    } catch (e) {
      throw new Error((e.response.data as IStatus).message);
    }
  }

  public static async delete(name: string, namespace: string = "default") {
    const url = this.getLink(namespace, name);
    return axios.delete(url);
  }

  public static async get(name: string, namespace: string = "default") {
    const url = this.getLink(namespace, name);
    const { data } = await axios.get<ISecret>(url);
    return data;
  }

  public static async list(namespace: string = "default") {
    const url = Secret.getLink(namespace);
    const { data } = await axios.get<ISecret>(url);
    return data;
  }

  private static getOwnerReference(owner: IResource) {
    return owner
      ? ({
          apiVersion: owner.apiVersion,
          blockOwnerDeletion: true,
          kind: owner.kind,
          name: owner.metadata.name,
          uid: owner.metadata.uid,
        } as IOwnerReference)
      : undefined;
  }

  private static getLink(namespace: string, name?: string): string {
    return `/api/kube/api/v1/namespaces/${namespace}/secrets${name ? `/${name}` : ""}`;
  }
}
