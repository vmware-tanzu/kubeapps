import { axiosWithAuth } from "./AxiosInstance";
import { IK8sList, ISecret } from "./types";
import * as url from "./url";

export default class Secret {
  public static async get(cluster: string, namespace: string, name: string) {
    const u = url.api.k8s.secret(cluster, namespace, name);
    const { data } = await axiosWithAuth.get<ISecret>(u);
    return data;
  }

  public static async list(cluster: string, namespace: string, fieldSelector?: string) {
    const u = url.api.k8s.secrets(cluster, namespace, fieldSelector);
    const { data } = await axiosWithAuth.get<IK8sList<ISecret, {}>>(u);
    return data;
  }

  public static async createPullSecret(
    cluster: string,
    name: string,
    user: string,
    password: string,
    email: string,
    server: string,
    namespace: string,
  ) {
    const u = url.api.k8s.secrets(cluster, namespace);
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
    const { data } = await axiosWithAuth.post<ISecret>(u, {
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
}
