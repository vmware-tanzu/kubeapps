import { axiosWithAuth } from "./AxiosInstance";
import { APIBase } from "./Kube";
import { IK8sList, IOwnerReference, ISecret } from "./types";

export default class Secret {
  public static async create(
    name: string,
    secrets: { [s: string]: string },
    owner: IOwnerReference | undefined,
    namespace: string,
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

  public static async delete(name: string, namespace: string) {
    const url = this.getLink(namespace, name);
    return axiosWithAuth.delete(url);
  }

  public static async get(name: string, namespace: string) {
    const url = this.getLink(namespace, name);
    const { data } = await axiosWithAuth.get<ISecret>(url);
    return data;
  }

  public static async list(namespace: string) {
    const url = Secret.getLink(namespace);
    const { data } = await axiosWithAuth.get<IK8sList<ISecret, {}>>(url);
    return data;
  }

  public static async createPullSecret(
    name: string,
    user: string,
    password: string,
    email: string,
    server: string,
    namespace: string,
  ) {
    const url = Secret.getLink(namespace);
    const dockercfg = {
      auths: {
        [server]: {
          username: user,
          password,
          email,
          auth: btoa(`${user}:${password}`),
        },
      },
    };
    const { data } = await axiosWithAuth.post<ISecret>(url, {
      apiVersion: "v1",
      stringData: {
        ".dockerconfigjson": JSON.stringify(dockercfg),
      },
      kind: "Secret",
      metadata: {
        name,
      },
      type: "kubernetes.io/dockerconfigjson",
    });
    return data;
  }

  private static getLink(namespace: string, name?: string): string {
    return `${APIBase}/api/v1/namespaces/${namespace}/secrets${name ? `/${name}` : ""}`;
  }
}
