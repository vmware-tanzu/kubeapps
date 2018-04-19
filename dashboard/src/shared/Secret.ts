import axios from "axios";

import { ISecret, IStatus } from "./types";

export default class Secret {
  public static async create(name: string, secrets: any, namespace: string = "default") {
    const url = Secret.getLink(namespace);
    try {
      const { data } = await axios.post<ISecret>(url, {
        apiVersion: "v1",
        data: secrets,
        kind: "Secret",
        metadata: { name },
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

  private static getLink(namespace: string, name?: string): string {
    return `/api/kube/api/v1/namespaces/${namespace}/secrets${name ? `/${name}` : ""}`;
  }
}
