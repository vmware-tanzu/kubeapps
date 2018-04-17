import axios from "axios";

import { ISecret, IStatus } from "./types";

export default class Secret {
  public static async create(name: string, secrets: any) {
    const url = Secret.getLink();
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

  public static async delete(name: string) {
    const url = this.getLink(name);
    return axios.delete(url);
  }

  public static async get(name: string) {
    const url = this.getLink(name);
    const { data } = await axios.get<ISecret>(url);
    return data;
  }

  public static async list() {
    const url = Secret.getLink();
    const { data } = await axios.get<ISecret>(url);
    return data;
  }

  private static getLink(name?: string): string {
    return `/api/kube/api/v1/namespaces/kubeapps/secrets${name ? `/${name}` : ""}`;
  }
}
