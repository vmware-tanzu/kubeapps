// Copyright 2018-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import {
  CreateSecretRequest,
  SecretType,
} from "gen/kubeappsapis/plugins/resources/v1alpha1/resources_pb";
import { KubeappsGrpcClient } from "./KubeappsGrpcClient";
import { convertGrpcAuthError } from "./utils";

export default class Secret {
  public static resourcesServiceClient = () =>
    new KubeappsGrpcClient().getResourcesServiceClientImpl();

  // TODO(agamez): unused method, remove?
  public static async getDockerConfigSecretNames(cluster: string, namespace: string) {
    const result = await this.resourcesServiceClient()
      .getSecretNames({
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
      if (type === SecretType.DOCKER_CONFIG_JSON) {
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
      .createSecret(
        new CreateSecretRequest({
          context: {
            cluster,
            namespace,
          },
          name,
          type: SecretType.DOCKER_CONFIG_JSON,
          stringData: {
            ".dockerconfigjson": JSON.stringify(dockercfg),
          },
        }),
      )
      .catch((e: any) => {
        throw convertGrpcAuthError(e);
      });
  }
}
