import { KubeappsGrpcClient } from "./KubeappsGrpcClient";
import {
  CreateSecretRequest,
  SecretType,
} from "gen/kubeappsapis/plugins/resources/v1alpha1/resources";

export default class Secret {
  public static resourcesClient = () => new KubeappsGrpcClient().getResourcesServiceClientImpl();

  public static async getDockerConfigSecretNames(cluster: string, namespace: string) {
    const result = await this.resourcesClient().GetSecretNames({
      context: {
        cluster,
        namespace,
      },
    });

    const secretNames = [];
    for (const [name, type] of Object.entries(result.secretNames)) {
      if (type === SecretType.SECRET_TYPE_DOCKER_CONFIG_JSON) {
        secretNames.push(name);
      }
    }
    return secretNames;
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
