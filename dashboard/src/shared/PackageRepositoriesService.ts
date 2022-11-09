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
import {
  GetConfiguredPluginsResponse,
  Plugin,
} from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import {
  HelmPackageRepositoryCustomDetail,
  protobufPackage as helmProtobufPackage,
  ProxyOptions,
} from "gen/kubeappsapis/plugins/helm/packages/v1alpha1/helm";
import {
  KappControllerPackageRepositoryCustomDetail,
  protobufPackage as kappControllerProtobufPackage,
} from "gen/kubeappsapis/plugins/kapp_controller/packages/v1alpha1/kapp_controller";
import {
  FluxPackageRepositoryCustomDetail,
  protobufPackage as fluxv2ProtobufPackage,
} from "gen/kubeappsapis/plugins/fluxv2/packages/v1alpha1/fluxv2";
import KubeappsGrpcClient from "./KubeappsGrpcClient";
import {
  IPkgRepoFormData,
  PluginNames,
  IResourcePermission,
  IPackageRepositoryPermission,
} from "./types";
import { convertGrpcAuthError } from "./utils";
import { Kube } from "./Kube";
import { IConfig } from "./Config";

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
          nodeSelector: {},
          tolerations: [],
          securityContext: {
            supplementalGroups: [],
          },
        } as HelmPackageRepositoryCustomDetail;

        // populate the imagesPullSecret if it's not empty
        if (
          detail?.imagesPullSecret?.secretRef ||
          Object.values(detail?.imagesPullSecret?.credentials as DockerCredentials).some(e => !!e)
        ) {
          helmCustomDetail.imagesPullSecret = detail.imagesPullSecret;
        }

        // populate the proxyOptions if it's not empty
        if (Object.values(detail?.proxyOptions as ProxyOptions).some(e => !!e)) {
          helmCustomDetail.proxyOptions = {
            enabled: detail.proxyOptions?.enabled || false,
            httpProxy: detail.proxyOptions?.httpProxy || "",
            httpsProxy: detail.proxyOptions?.httpsProxy || "",
            noProxy: detail.proxyOptions?.noProxy || "",
          };
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
      case PluginNames.PACKAGES_FLUX:
        return {
          typeUrl: `${fluxv2ProtobufPackage}.FluxPackageRepositoryCustomDetail`,
          value: FluxPackageRepositoryCustomDetail.encode(
            request.customDetail as FluxPackageRepositoryCustomDetail,
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

  public static getRepositoriesPermissions = (
    namespace: string,
    config: IConfig,
    plugin: Plugin,
  ): IPackageRepositoryPermission | undefined => {
    switch (plugin.name) {
      case PluginNames.PACKAGES_HELM:
        return {
          global: this.getResourcePermissions(
            config.kubeappsCluster,
            config.helmGlobalNamespace,
            "kubeapps.com",
            "apprepositories",
          ),
          namespaced: this.getResourcePermissions(
            config.kubeappsCluster,
            namespace,
            "kubeapps.com",
            "apprepositories",
          ),
          plugin: plugin,
        };
      case PluginNames.PACKAGES_KAPP:
        return {
          global: this.getResourcePermissions(
            config.kubeappsCluster,
            config.carvelGlobalNamespace,
            "packaging.carvel.dev",
            "packagerepositories",
          ),
          namespaced: this.getResourcePermissions(
            config.kubeappsCluster,
            namespace,
            "packaging.carvel.dev",
            "packagerepositories",
          ),
          plugin: plugin,
        };
      case PluginNames.PACKAGES_FLUX:
        return {
          namespaced: this.getResourcePermissions(
            config.kubeappsCluster,
            namespace,
            "source.toolkit.fluxcd.io",
            "helmrepositories",
          ),
          plugin: plugin,
        };
    }
    return undefined;
  };

  // TODO(castelblanque) Provide the repository resource GVR from the backend plugin
  // and make a single canI call that groups all verbs
  private static getResourcePermissions = (
    cluster: string,
    namespace: string,
    apiGroup: string,
    resource: string,
  ): IResourcePermission => {
    const permissions = {
      namespace: namespace,
    } as IResourcePermission;
    Promise.allSettled([
      Kube.canI(cluster, apiGroup, resource, "list", namespace)
        .then(allowed => (permissions.list = allowed))
        ?.catch(() => (permissions.list = false)),
      Kube.canI(cluster, apiGroup, resource, "update", namespace)
        .then(allowed => (permissions.update = allowed))
        ?.catch(() => (permissions.update = false)),
      Kube.canI(cluster, apiGroup, resource, "create", namespace)
        .then(allowed => (permissions.create = allowed))
        ?.catch(() => (permissions.create = false)),
      Kube.canI(cluster, apiGroup, resource, "delete", namespace)
        .then(allowed => (permissions.delete = allowed))
        ?.catch(() => (permissions.delete = false)),
    ]);
    return permissions;
  };
}
