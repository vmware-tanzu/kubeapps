// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import {
  CreateSecretRequest,
  SecretType,
} from "gen/kubeappsapis/plugins/resources/v1alpha1/resources";
import { KubeappsGrpcClient } from "./KubeappsGrpcClient";
import { convertGrpcAuthError } from "./utils";

export default class Secret {
  public static resourcesServiceClient = () =>
    new KubeappsGrpcClient().getResourcesServiceClientImpl();

  // TODO(agamez): unused method, remove?
  public static async getDockerConfigSecretNames(cluster: string, namespace: string) {
    const result = await this.resourcesServiceClient()
      .GetSecretNames({
        context: {
          cluster,
          namespace,
        },
      })
      .catch((e: any) => {
        throw convertGrpcAuthError(e);
      });

    const secretNames = [];
    for (const [name, type] of Object.entries(result.secretNames)) {
      if (type === SecretType.SECRET_TYPE_DOCKER_CONFIG_JSON) {
        secretNames.push(name);
      }
    }
    return secretNames;
  }

  // TODO(agamez): unused method, remove?
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
          auth: Buffer.from(`${user}:${password}`).toString("base64"),
        },
      },
    };
    await this.resourcesServiceClient()
      .CreateSecret({
        context: {
          cluster,
          namespace,
        },
        name,
        type: SecretType.SECRET_TYPE_DOCKER_CONFIG_JSON,
        stringData: {
          ".dockerconfigjson": JSON.stringify(dockercfg),
        },
      } as CreateSecretRequest)
      .catch((e: any) => {
        throw convertGrpcAuthError(e);
      });
  }
}
