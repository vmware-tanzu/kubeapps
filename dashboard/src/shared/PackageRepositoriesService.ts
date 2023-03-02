// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { Any } from "gen/google/protobuf/any";
import { Context } from "gen/kubeappsapis/core/packages/v1alpha1/packages_pb";
import {
  AddPackageRepositoryRequest,
  DeletePackageRepositoryResponse,
  DockerCredentials,
  GetPackageRepositoryDetailResponse,
  GetPackageRepositorySummariesResponse,
  OpaqueCredentials,
  PackageRepositoryAuth,
  PackageRepositoryAuth_PackageRepositoryAuthType,
  PackageRepositoryReference,
  PackageRepositoryTlsConfig,
  SecretKeyReference,
  SshCredentials,
  TlsCertKey,
  UpdatePackageRepositoryRequest,
  UsernamePassword,
} from "gen/kubeappsapis/core/packages/v1alpha1/repositories_pb";
import { GetConfiguredPluginsResponse } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins_pb";
import {
  HelmPackageRepositoryCustomDetail,
  protobufPackage as helmProtobufPackage,
  ProxyOptions,
} from "gen/kubeappsapis/plugins/helm/packages/v1alpha1/helm_pb";
import {
  KappControllerPackageRepositoryCustomDetail,
  protobufPackage as kappControllerProtobufPackage,
} from "gen/kubeappsapis/plugins/kapp_controller/packages/v1alpha1/kapp_controller_pb";
import {
  FluxPackageRepositoryCustomDetail,
  protobufPackage as fluxv2ProtobufPackage,
} from "gen/kubeappsapis/plugins/fluxv2/packages/v1alpha1/fluxv2_pb";
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
      .getPackageRepositorySummaries({ context })
      .catch((e: any) => {
        throw convertGrpcAuthError(e);
      });
  }

  public static async getPackageRepositoryDetail(
    packageRepoRef: PackageRepositoryReference,
  ): Promise<GetPackageRepositoryDetailResponse> {
    return await this.coreRepositoriesClient()
      .getPackageRepositoryDetail({ packageRepoRef })
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
      .addPackageRepository(addPackageRepositoryRequest)
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
      .updatePackageRepository(updatePackageRepositoryRequest)
      .catch((e: any) => {
        throw convertGrpcAuthError(e);
      });
  }

  public static async deletePackageRepository(
    packageRepoRef: PackageRepositoryReference,
  ): Promise<DeletePackageRepositoryResponse> {
    return await this.coreRepositoriesClient()
      .deletePackageRepository({
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

    // auth type
    if (request.authMethod) {
      addPackageRepositoryRequest.auth = {
        ...addPackageRepositoryRequest.auth,
        type: request.authMethod,
      } as PackageRepositoryAuth;
    }

    // auth/tls - user entered
    if (!request.isUserManaged) {
      switch (request.authMethod) {
        case (PackageRepositoryAuth_PackageRepositoryAuthType.AUTHORIZATION_HEADER,
          PackageRepositoryAuth_PackageRepositoryAuthType.BEARER):
          if (request.authHeader) {
            addPackageRepositoryRequest.auth = {
              ...addPackageRepositoryRequest.auth,
              header: request.authHeader,
            } as PackageRepositoryAuth;
          }
          break;
        case PackageRepositoryAuth_PackageRepositoryAuthType.BASIC_AUTH:
          if (Object.values(request.basicAuth).some(e => !!e)) {
            addPackageRepositoryRequest.auth = {
              ...addPackageRepositoryRequest.auth,
              usernamePassword: {
                username: request.basicAuth.username,
                password: request.basicAuth.password,
              } as UsernamePassword,
            } as PackageRepositoryAuth;
          }
          break;
        case PackageRepositoryAuth_PackageRepositoryAuthType.DOCKER_CONFIG_JSON:
          if (Object.values(request.dockerRegCreds).some(e => !!e)) {
            addPackageRepositoryRequest.auth = {
              ...addPackageRepositoryRequest.auth,
              dockerCreds: { ...request.dockerRegCreds } as DockerCredentials,
            } as PackageRepositoryAuth;
          }
          break;
        case PackageRepositoryAuth_PackageRepositoryAuthType.SSH:
          if (Object.values(request.sshCreds).some(e => !!e)) {
            addPackageRepositoryRequest.auth = {
              ...addPackageRepositoryRequest.auth,
              sshCreds: {
                ...request.sshCreds,
              } as SshCredentials,
            } as PackageRepositoryAuth;
          }
          break;
        case PackageRepositoryAuth_PackageRepositoryAuthType.TLS:
          if (Object.values(request.tlsCertKey).some(e => !!e)) {
            addPackageRepositoryRequest.auth = {
              ...addPackageRepositoryRequest.auth,
              tlsCertKey: { ...request.tlsCertKey } as TlsCertKey,
            } as PackageRepositoryAuth;
          }
          break;
        case PackageRepositoryAuth_PackageRepositoryAuthType.OPAQUE:
          if (Object.values(request.opaqueCreds.data).some(e => !!e)) {
            addPackageRepositoryRequest.auth = {
              ...addPackageRepositoryRequest.auth,
              opaqueCreds: { ...request.opaqueCreds } as OpaqueCredentials,
            } as PackageRepositoryAuth;
          }
          break;
      }

      if (request.customCA) {
        addPackageRepositoryRequest.tlsConfig = {
          ...addPackageRepositoryRequest.tlsConfig,
          certAuthority: request.customCA,
        } as PackageRepositoryTlsConfig;
      }
    }

    // auth/tls - user managed
    if (request.isUserManaged) {
      if (request.secretTLSName) {
        addPackageRepositoryRequest.tlsConfig = {
          ...addPackageRepositoryRequest.tlsConfig,
          secretRef: {
            name: request.secretTLSName,
          } as SecretKeyReference,
        } as PackageRepositoryTlsConfig;
      }
      if (
        request.authMethod !==
        PackageRepositoryAuth_PackageRepositoryAuthType.UNSPECIFIED &&
        request.secretAuthName
      ) {
        addPackageRepositoryRequest.auth = {
          ...addPackageRepositoryRequest.auth,
          secretRef: {
            name: request.secretAuthName,
          } as SecretKeyReference,
        } as PackageRepositoryAuth;
      }
    }

    // misc fields
    if (request.passCredentials) {
      addPackageRepositoryRequest.auth = {
        ...addPackageRepositoryRequest.auth,
        passCredentials: request.passCredentials,
      } as PackageRepositoryAuth;
    }
    if (request.skipTLS) {
      addPackageRepositoryRequest.tlsConfig = {
        ...addPackageRepositoryRequest.tlsConfig,
        insecureSkipVerify: request.skipTLS,
      } as PackageRepositoryTlsConfig;
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
          (detail?.imagesPullSecret?.credentials &&
            Object.values((detail?.imagesPullSecret?.credentials || {}) as DockerCredentials).some(
              e => !!e,
            ))
        ) {
          helmCustomDetail.imagesPullSecret = detail.imagesPullSecret;
        }

        // populate the proxyOptions if it's not empty
        if (Object.values((detail?.proxyOptions || {}) as ProxyOptions).some(e => !!e)) {
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
      .getConfiguredPlugins({})
      .catch((e: any) => {
        throw convertGrpcAuthError(e);
      });
  }

  public static async getRepositoriesPermissions(cluster: string, namespace: string) {
    const resp = await this.coreRepositoriesClient().getPackageRepositoryPermissions({
      context: {
        cluster: cluster,
        namespace: namespace,
      },
    });
    return resp.permissions;
  }
}
