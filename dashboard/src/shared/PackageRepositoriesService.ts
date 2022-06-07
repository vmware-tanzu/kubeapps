// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { Any } from "gen/google/protobuf/any";
import { Context } from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import {
  AddPackageRepositoryRequest,
  DeletePackageRepositoryResponse,
  DockerCredentials,
  GetPackageRepositoryDetailResponse,
  GetPackageRepositorySummariesResponse,
  PackageRepositoryAuth,
  PackageRepositoryReference,
  PackageRepositoryTlsConfig,
  SecretKeyReference,
  UpdatePackageRepositoryRequest,
  UsernamePassword,
} from "gen/kubeappsapis/core/packages/v1alpha1/repositories";
import { RepositoryCustomDetails } from "gen/kubeappsapis/plugins/helm/packages/v1alpha1/helm";
import KubeappsGrpcClient from "./KubeappsGrpcClient";
import { IPkgRepoFormData } from "./types";
import { PluginNames } from "./utils";

export class PackageRepositoriesService {
  public static coreRepositoriesClient = () =>
    new KubeappsGrpcClient().getRepositoriesServiceClientImpl();
  public static helmRepositoriesClient = () =>
    new KubeappsGrpcClient().getHelmRepositoriesServiceClientImpl();

  public static async getPackageRepositorySummaries(
    context: Context,
  ): Promise<GetPackageRepositorySummariesResponse> {
    return await this.coreRepositoriesClient().GetPackageRepositorySummaries({ context });
  }

  public static async getPackageRepositoryDetail(
    packageRepoRef: PackageRepositoryReference,
  ): Promise<GetPackageRepositoryDetailResponse> {
    // since the Helm plugin has its own fields (ociRepositories, filter),
    // we invoke it directly instead of using kthe core API client.
    switch (packageRepoRef?.plugin?.name) {
      case PluginNames.PACKAGES_HELM:
        return await this.helmRepositoriesClient().GetPackageRepositoryDetail({ packageRepoRef });
      default:
        return await this.coreRepositoriesClient().GetPackageRepositoryDetail({ packageRepoRef });
    }
  }

  public static async addPackageRepository(
    cluster: string,
    namespace: string,
    request: IPkgRepoFormData,
    namespaceScoped: boolean,
  ) {
    const addPackageRepositoryRequest = this.buildAddOrUpdateRequest(
      false,
      cluster,
      namespace,
      request,
      namespaceScoped,
    );

    // since the Helm plugin has its own fields (ociRepositories, filter),
    // we invoke it directly instead of using kthe core API client.
    switch (request.plugin.name) {
      case PluginNames.PACKAGES_HELM:
        console.log(request);
        console.log(addPackageRepositoryRequest);
        return await this.helmRepositoriesClient().AddPackageRepository(
          addPackageRepositoryRequest,
        );
      default:
        return await this.coreRepositoriesClient().AddPackageRepository(
          addPackageRepositoryRequest,
        );
    }
  }

  public static async updatePackageRepository(
    cluster: string,
    namespace: string,
    request: IPkgRepoFormData,
  ) {
    const updatePackageRepositoryRequest = this.buildAddOrUpdateRequest(
      true,
      cluster,
      namespace,
      request,
      undefined,
    );

    // since the Helm plugin has its own fields (ociRepositories, filter),
    // we invoke it directly instead of using kthe core API client.
    switch (request.plugin.name) {
      case PluginNames.PACKAGES_HELM:
        return await this.helmRepositoriesClient().UpdatePackageRepository(
          updatePackageRepositoryRequest,
        );
      default:
        return await this.coreRepositoriesClient().UpdatePackageRepository(
          updatePackageRepositoryRequest,
        );
    }
  }

  public static async deletePackageRepository(
    packageRepoRef: PackageRepositoryReference,
  ): Promise<DeletePackageRepositoryResponse> {
    return await this.coreRepositoriesClient().DeletePackageRepository({
      packageRepoRef,
    });
  }

  private static buildAddOrUpdateRequest(
    isUpdate: boolean,
    cluster: string,
    namespace: string,
    request: IPkgRepoFormData,
    namespaceScoped?: boolean,
  ) {
    const addPackageRepositoryRequest = {
      context: { cluster, namespace },
      name: request.name,
      description: request.description,
      namespaceScoped: namespaceScoped,
      type: request.type,
      url: request.url,
      interval: request.interval,
      plugin: request.plugin,
    } as AddPackageRepositoryRequest;

    // add optional fields if present in the request
    if (request.authHeader) {
      addPackageRepositoryRequest.auth = {
        ...addPackageRepositoryRequest.auth,
        header: request.authHeader,
      } as PackageRepositoryAuth;
    }
    if (request.passCredentials) {
      addPackageRepositoryRequest.auth = {
        ...addPackageRepositoryRequest.auth,
        passCredentials: request.passCredentials,
      } as PackageRepositoryAuth;
    }
    if (request.authMethod) {
      addPackageRepositoryRequest.auth = {
        ...addPackageRepositoryRequest.auth,
        type: request.authMethod,
      } as PackageRepositoryAuth;
    }
    if (Object.values(request.basicAuth).some(e => !!e)) {
      addPackageRepositoryRequest.auth = {
        ...addPackageRepositoryRequest.auth,
        usernamePassword: {
          username: request.basicAuth.username,
          password: request.basicAuth.password,
        } as UsernamePassword,
      } as PackageRepositoryAuth;
    }
    if (Object.values(request.dockerRegCreds).some(e => !!e)) {
      addPackageRepositoryRequest.auth = {
        ...addPackageRepositoryRequest.auth,
        dockerCreds: { ...request.dockerRegCreds } as DockerCredentials,
      } as PackageRepositoryAuth;
    }
    if (request.customCA) {
      addPackageRepositoryRequest.tlsConfig = {
        ...addPackageRepositoryRequest.tlsConfig,
        certAuthority: request.customCA,
      } as PackageRepositoryTlsConfig;
    }
    if (request.skipTLS) {
      addPackageRepositoryRequest.tlsConfig = {
        ...addPackageRepositoryRequest.tlsConfig,
        insecureSkipVerify: request.skipTLS,
      } as PackageRepositoryTlsConfig;
    }
    if (request.secretTLSName) {
      addPackageRepositoryRequest.tlsConfig = {
        ...addPackageRepositoryRequest.tlsConfig,
        secretRef: {
          name: request.secretTLSName,
        } as SecretKeyReference,
      } as PackageRepositoryTlsConfig;
    }
    if (request.secretAuthName) {
      addPackageRepositoryRequest.auth = {
        ...addPackageRepositoryRequest.auth,
        secretRef: {
          name: request.secretAuthName,
        } as SecretKeyReference,
      } as PackageRepositoryAuth;
    }
    // if using the Helm plugin, add its custom fields.
    // An "Any" object has  "typeUrl" with the FQN of the type and a "value",
    // which is the result of the encoding (+finish(), to get the Uint8Array)
    // of the actual custom object
    if (request.plugin?.name === PluginNames.PACKAGES_HELM) {
      addPackageRepositoryRequest.customDetail = {
        typeUrl: "kubeappsapis.plugins.helm.packages.v1alpha1.RepositoryCustomDetails",
        value: RepositoryCustomDetails.encode({
          dockerRegistrySecrets: request.customDetails.dockerRegistrySecrets,
          ociRepositories: request.customDetails.ociRepositories,
          filterRule: request.customDetails.filterRule,
          performValidation: request.customDetails.performValidation,
        } as RepositoryCustomDetails).finish(),
      } as Any;
    }

    if (isUpdate) {
      const updatePackageRepositoryRequest: UpdatePackageRepositoryRequest = {
        description: addPackageRepositoryRequest.description,
        interval: addPackageRepositoryRequest.interval,
        url: addPackageRepositoryRequest.url,
        auth: addPackageRepositoryRequest.auth,
        customDetail: addPackageRepositoryRequest.customDetail,
        tlsConfig: addPackageRepositoryRequest.tlsConfig,
        packageRepoRef: {
          identifier: addPackageRepositoryRequest.name,
          context: addPackageRepositoryRequest.context,
          plugin: addPackageRepositoryRequest.plugin,
        },
      };
      return updatePackageRepositoryRequest;
    }
    return addPackageRepositoryRequest as UpdatePackageRepositoryRequest;

    // TODO(agamez): -- currently unsupported configuration --
    // tlsConfig.secretRef={ key: "", name: "" }, // reference a secret to pass the CA certificate
    // auth.tlsCertKey: { cert: "", key: ""},  // cert and key for tls auth
  }
}
