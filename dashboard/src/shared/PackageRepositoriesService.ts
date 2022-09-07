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
  OpaqueCredentials,
  PackageRepositoryAuth,
  PackageRepositoryReference,
  PackageRepositoryTlsConfig,
  SecretKeyReference,
  SshCredentials,
  TlsCertKey,
  UpdatePackageRepositoryRequest,
  UsernamePassword,
} from "gen/kubeappsapis/core/packages/v1alpha1/repositories";
import { GetConfiguredPluginsResponse } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import {
  HelmPackageRepositoryCustomDetail,
  protobufPackage as helmProtobufPackage,
} from "gen/kubeappsapis/plugins/helm/packages/v1alpha1/helm";
import {
  KappControllerPackageRepositoryCustomDetail,
  protobufPackage as kappControllerProtobufPackage,
} from "gen/kubeappsapis/plugins/kapp_controller/packages/v1alpha1/kapp_controller";
import KubeappsGrpcClient from "./KubeappsGrpcClient";
import { IPkgRepoFormData, PluginNames } from "./types";
import { convertGrpcAuthError } from "./utils";

export class PackageRepositoriesService {
  public static coreRepositoriesClient = () =>
    new KubeappsGrpcClient().getRepositoriesServiceClientImpl();
  public static pluginsServiceClientImpl = () =>
    new KubeappsGrpcClient().getPluginsServiceClientImpl();

  public static async getPackageRepositorySummaries(
    context: Context,
  ): Promise<GetPackageRepositorySummariesResponse> {
    return await this.coreRepositoriesClient()
      .GetPackageRepositorySummaries({ context })
      .catch((e: any) => {
        throw convertGrpcAuthError(e);
      });
  }

  public static async getPackageRepositoryDetail(
    packageRepoRef: PackageRepositoryReference,
  ): Promise<GetPackageRepositoryDetailResponse> {
    return await this.coreRepositoriesClient()
      .GetPackageRepositoryDetail({ packageRepoRef })
      .catch((e: any) => {
        throw convertGrpcAuthError(e);
      });
  }

  public static async addPackageRepository(cluster: string, request: IPkgRepoFormData) {
    const addPackageRepositoryRequest = PackageRepositoriesService.buildAddOrUpdateRequest(
      false,
      cluster,
      request,
      PackageRepositoriesService.buildEncodedCustomDetail(request),
    );

    return await this.coreRepositoriesClient()
      .AddPackageRepository(addPackageRepositoryRequest)
      .catch((e: any) => {
        throw convertGrpcAuthError(e);
      });
  }

  public static async updatePackageRepository(cluster: string, request: IPkgRepoFormData) {
    const updatePackageRepositoryRequest = PackageRepositoriesService.buildAddOrUpdateRequest(
      true,
      cluster,
      request,
      PackageRepositoriesService.buildEncodedCustomDetail(request),
    );

    return await this.coreRepositoriesClient()
      .UpdatePackageRepository(updatePackageRepositoryRequest)
      .catch((e: any) => {
        throw convertGrpcAuthError(e);
      });
  }

  public static async deletePackageRepository(
    packageRepoRef: PackageRepositoryReference,
  ): Promise<DeletePackageRepositoryResponse> {
    return await this.coreRepositoriesClient()
      .DeletePackageRepository({
        packageRepoRef,
      })
      .catch((e: any) => {
        throw convertGrpcAuthError(e);
      });
  }

  private static buildAddOrUpdateRequest(
    isUpdate: boolean,
    cluster: string,
    request: IPkgRepoFormData,
    pluginCustomDetail?: any,
  ) {
    const addPackageRepositoryRequest = {
      context: { cluster, namespace: request.namespace },
      name: request.name,
      description: request.description,
      namespaceScoped: request.isNamespaceScoped,
      type: request.type,
      url: request.url,
      interval: request.interval,
      plugin: request.plugin,
      customDetail: pluginCustomDetail,
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
    if (Object.values(request.sshCreds).some(e => !!e)) {
      addPackageRepositoryRequest.auth = {
        ...addPackageRepositoryRequest.auth,
        sshCreds: {
          ...request.sshCreds,
        } as SshCredentials,
      } as PackageRepositoryAuth;
    }
    if (Object.values(request.tlsCertKey).some(e => !!e)) {
      addPackageRepositoryRequest.auth = {
        ...addPackageRepositoryRequest.auth,
        tlsCertKey: { ...request.tlsCertKey } as TlsCertKey,
      } as PackageRepositoryAuth;
    }
    if (Object.values(request.opaqueCreds.data).some(e => !!e)) {
      addPackageRepositoryRequest.auth = {
        ...addPackageRepositoryRequest.auth,
        opaqueCreds: { ...request.opaqueCreds } as OpaqueCredentials,
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
    return addPackageRepositoryRequest;
  }

  private static buildEncodedCustomDetail(request: IPkgRepoFormData) {
    // if using a plugin with customDetail, encode its custom fields,
    // otherwise skip it
    if (!request.customDetail) {
      return;
    }
    // An "Any" object has "typeUrl" with the FQN of the type and a "value",
    // which is the result of the encoding (+finish(), to get the Uint8Array)
    // of the actual custom object
    switch (request.plugin?.name) {
      case PluginNames.PACKAGES_HELM: {
        const detail = request.customDetail as HelmPackageRepositoryCustomDetail;
        const helmCustomDetail = {
          // populate the non-optional fields
          ociRepositories: detail?.ociRepositories || [],
          performValidation: !!detail?.performValidation,
          filterRule: detail?.filterRule,
        } as HelmPackageRepositoryCustomDetail;

        // populate the imagesPullSecret if it's not empty
        if (
          detail?.imagesPullSecret?.secretRef ||
          Object.values(detail?.imagesPullSecret?.credentials as DockerCredentials).some(e => !!e)
        ) {
          helmCustomDetail.imagesPullSecret = detail.imagesPullSecret;
        }

        return {
          typeUrl: `${helmProtobufPackage}.HelmPackageRepositoryCustomDetail`,
          value: HelmPackageRepositoryCustomDetail.encode(helmCustomDetail).finish(),
        } as Any;
      }
      case PluginNames.PACKAGES_KAPP:
        return {
          typeUrl: `${kappControllerProtobufPackage}.KappControllerPackageRepositoryCustomDetail`,
          value: KappControllerPackageRepositoryCustomDetail.encode(
            request.customDetail as KappControllerPackageRepositoryCustomDetail,
          ).finish(),
        } as Any;
      default:
        return;
    }
  }

  public static async getConfiguredPlugins(): Promise<GetConfiguredPluginsResponse> {
    return await this.pluginsServiceClientImpl()
      .GetConfiguredPlugins({})
      .catch((e: any) => {
        throw convertGrpcAuthError(e);
      });
  }
}
