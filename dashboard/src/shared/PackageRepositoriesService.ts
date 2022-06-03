// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { Any } from "gen/google/protobuf/any";
import {
  AddPackageRepositoryRequest,
  DeletePackageRepositoryResponse,
  GetPackageRepositoryDetailResponse,
  GetPackageRepositorySummariesResponse,
  PackageRepositoryAuth,
  PackageRepositoryAuth_PackageRepositoryAuthType,
  PackageRepositoryReference,
  PackageRepositoryTlsConfig,
  SecretKeyReference,
  UpdatePackageRepositoryRequest,
  UsernamePassword,
} from "gen/kubeappsapis/core/packages/v1alpha1/repositories";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import { RepositoryCustomDetails } from "gen/kubeappsapis/plugins/helm/packages/v1alpha1/helm";
import { axiosWithAuth } from "./AxiosInstance";
import KubeappsGrpcClient from "./KubeappsGrpcClient";
import { IAppRepositoryFilter } from "./types";
import * as url from "./url";
import { PluginNames } from "./utils";

export class PackageRepositoriesService {
  public static coreRepositoriesClient = () =>
    new KubeappsGrpcClient().getRepositoriesServiceClientImpl();
  public static helmRepositoriesClient = () =>
    new KubeappsGrpcClient().getHelmRepositoriesServiceClientImpl();

  public static async getPackageRepositorySummaries(
    cluster: string,
    namespace?: string,
  ): Promise<GetPackageRepositorySummariesResponse> {
    return await this.coreRepositoriesClient().GetPackageRepositorySummaries({
      context: {
        cluster,
        namespace,
      },
    });
  }

  public static async getPackageRepositoryDetail(
    packageRepoRef: PackageRepositoryReference,
  ): Promise<GetPackageRepositoryDetailResponse> {
    return await this.coreRepositoriesClient().GetPackageRepositoryDetail({ packageRepoRef });
  }

  public static async addPackageRepository(
    cluster: string,
    name: string,
    plugin: Plugin,
    namespace: string,
    repoURL: string,
    type: string,
    description: string,
    authHeader: string,
    authRegCreds: string,
    customCA: string,
    syncJobPodTemplate: any,
    registrySecrets: string[],
    ociRepositories: string[],
    skipTLS: boolean,
    passCredentials: boolean,
    namespaceScoped: boolean,
    authMethod: PackageRepositoryAuth_PackageRepositoryAuthType,
    interval: number,
    username: string,
    password: string,
    dryRun: boolean,
    filter?: IAppRepositoryFilter,
  ) {
    const addPackageRepositoryRequest = this.buildAddOrUpdateRequest(
      false,
      cluster,
      name,
      plugin,
      namespace,
      repoURL,
      type,
      description,
      authHeader,
      authRegCreds,
      customCA,
      syncJobPodTemplate,
      registrySecrets,
      ociRepositories,
      skipTLS,
      passCredentials,
      authMethod,
      interval,
      username,
      password,
      dryRun,
      filter,
      namespaceScoped,
    );

    // since the Helm plugin has its own fields (syncJobPodTemplate, ociRepositories, filter),
    // we invoke it directly instead of using kthe core API client.
    if (plugin.name === PluginNames.PACKAGES_HELM) {
      return await this.helmRepositoriesClient().AddPackageRepository(addPackageRepositoryRequest);
    } else {
      return await this.coreRepositoriesClient().AddPackageRepository(addPackageRepositoryRequest);
    }
  }

  public static async updatePackageRepository(
    cluster: string,
    name: string,
    plugin: Plugin,
    namespace: string,
    repoURL: string,
    type: string,
    description: string,
    authHeader: string,
    authRegCreds: string,
    customCA: string,
    syncJobPodTemplate: any,
    registrySecrets: string[],
    ociRepositories: string[],
    skipTLS: boolean,
    passCredentials: boolean,
    authMethod: PackageRepositoryAuth_PackageRepositoryAuthType,
    interval: number,
    username: string,
    password: string,
    performValidation: boolean,
    filter?: IAppRepositoryFilter,
  ) {
    const updatePackageRepositoryRequest = this.buildAddOrUpdateRequest(
      true,
      cluster,
      name,
      plugin,
      namespace,
      repoURL,
      type,
      description,
      authHeader,
      authRegCreds,
      customCA,
      syncJobPodTemplate,
      registrySecrets,
      ociRepositories,
      skipTLS,
      passCredentials,
      authMethod,
      interval,
      username,
      password,
      performValidation,
      filter,
      undefined,
    );

    // since the Helm plugin has its own fields (syncJobPodTemplate, ociRepositories, filter),
    // we invoke it directly instead of using kthe core API client.
    if (plugin.name === PluginNames.PACKAGES_HELM) {
      return await this.helmRepositoriesClient().UpdatePackageRepository(
        updatePackageRepositoryRequest,
      );
    } else {
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
    name: string,
    plugin: Plugin,
    namespace: string,
    repoURL: string,
    type: string,
    description: string,
    authHeader: string,
    authRegCreds: string,
    customCA: string,
    _syncJobPodTemplate: any,
    registrySecrets: string[],
    ociRepositories: string[],
    skipTLS: boolean,
    passCredentials: boolean,
    authMethod: PackageRepositoryAuth_PackageRepositoryAuthType,
    interval: number,
    username: string,
    password: string,
    performValidation: boolean,
    filter?: IAppRepositoryFilter,
    namespaceScoped?: boolean,
  ) {
    const addPackageRepositoryRequest = {
      context: { cluster, namespace },
      name,
      description,
      namespaceScoped,
      type,
      url: repoURL,
      interval,
      plugin,
    } as AddPackageRepositoryRequest;

    // add optional fields if present in the request
    if (authHeader) {
      addPackageRepositoryRequest.auth = {
        ...addPackageRepositoryRequest.auth,
        header: authHeader,
      } as PackageRepositoryAuth;
    }
    if (passCredentials) {
      addPackageRepositoryRequest.auth = {
        ...addPackageRepositoryRequest.auth,
        passCredentials: passCredentials,
      } as PackageRepositoryAuth;
    }
    if (authMethod) {
      addPackageRepositoryRequest.auth = {
        ...addPackageRepositoryRequest.auth,
        type: authMethod,
      } as PackageRepositoryAuth;
    }
    if (username || password) {
      addPackageRepositoryRequest.auth = {
        ...addPackageRepositoryRequest.auth,
        usernamePassword: {
          username,
          password,
        } as UsernamePassword,
      } as PackageRepositoryAuth;
    }
    if (customCA) {
      addPackageRepositoryRequest.tlsConfig = {
        ...addPackageRepositoryRequest.tlsConfig,
        certAuthority: customCA,
      } as PackageRepositoryTlsConfig;
    }
    if (skipTLS) {
      addPackageRepositoryRequest.tlsConfig = {
        ...addPackageRepositoryRequest.tlsConfig,
        insecureSkipVerify: skipTLS,
      } as PackageRepositoryTlsConfig;
    }
    if (
      authMethod ===
        PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON &&
      registrySecrets[0]
    ) {
      addPackageRepositoryRequest.tlsConfig = {
        ...addPackageRepositoryRequest.tlsConfig,
        secretRef: {
          name: registrySecrets[0],
        } as SecretKeyReference,
      } as PackageRepositoryTlsConfig;
    }
    if (authRegCreds) {
      addPackageRepositoryRequest.auth!.secretRef = {
        name: authRegCreds,
      } as SecretKeyReference;
    }

    // if using the Helm plugin, add its custom fields.
    // An "Any" object has  "typeUrl" with the FQN of the type and a "value",
    // which is the result of the encoding (+finish(), to get the Uint8Array)
    // of the actual custom object
    if (plugin?.name === PluginNames.PACKAGES_HELM) {
      addPackageRepositoryRequest.customDetail = {
        typeUrl: "kubeappsapis.plugins.helm.packages.v1alpha1.RepositoryCustomDetails",
        value: RepositoryCustomDetails.encode({
          dockerRegistrySecrets: registrySecrets,
          ociRepositories: ociRepositories,
          filterRule: filter,
          performValidation: performValidation,
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

    // auth.dockerCreds: { email: "", password: "", server: "", username: ""} // username and password for docker auth

    // tlsConfig.secretRef={ key: "", name: "" }, // reference a secret to pass the CA certificate

    // auth.tlsCertKey: { cert: "", key: ""},  // cert and key for tls auth
  }

  // ............................... DEPRECATED ...............................

  public static async getSecretForRepo(cluster: string, namespace: string, name: string) {
    const {
      data: { secret },
    } = await axiosWithAuth.get<any>(url.backend.apprepositories.get(cluster, namespace, name));
    return secret;
  }

  // TODO(agamez): the refresh functionallity is currently not implemented/supported in the new Repositories API. Decide whether removing it or not
  // public static async resync(cluster: string, namespace: string, name: string) {
  //   const { data } = await axiosWithAuth.post(
  //     url.backend.apprepositories.refresh(cluster, namespace, name),
  //     null,
  //   );
  //   return data;
}
