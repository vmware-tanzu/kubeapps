import { axiosWithAuth } from "./AxiosInstance";
import { IK8sList, ISecret } from "./types";
import * as url from "./url";
import { KubeappsGrpcClient } from "./KubeappsGrpcClient";
import {
  CreateSecretRequest,
  SecretType,
} from "gen/kubeappsapis/plugins/resources/v1alpha1/resources";

export default class Secret {
  public static resourcesClient = () => new KubeappsGrpcClient().getResourcesServiceClientImpl();
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
    await this.resourcesClient().CreateSecret({
      context: {
        cluster,
        namespace,
      },
      name,
      type: SecretType.SECRET_TYPE_DOCKER_CONFIG_JSON,
      stringData: {
        ".dockerconfigjson": JSON.stringify(dockercfg),
      },
    } as CreateSecretRequest);
  }
}
