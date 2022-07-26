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
  HelmPackageRepositoryCustomDetail,
  protobufPackage as helmProtobufPackage,
} from "gen/kubeappsapis/plugins/helm/packages/v1alpha1/helm";
import {
  KappControllerPackageRepositoryCustomDetail,
  protobufPackage as kappControllerProtobufPackage,
} from "gen/kubeappsapis/plugins/kapp_controller/packages/v1alpha1/kapp_controller";
import KubeappsGrpcClient from "./KubeappsGrpcClient";
import { IPkgRepoFormData } from "./types";
import { PluginNames } from "./utils";

export class PackageRepositoriesService {
  public static coreRepositoriesClient = () =>
    new KubeappsGrpcClient().getRepositoriesServiceClientImpl();

  public static async getPackageRepositorySummaries(
    context: Context,
  ): Promise<GetPackageRepositorySummariesResponse> {
    return await this.coreRepositoriesClient().GetPackageRepositorySummaries({ context });
  }

  public static async getPackageRepositoryDetail(
    packageRepoRef: PackageRepositoryReference,
  ): Promise<GetPackageRepositoryDetailResponse> {
    return await this.coreRepositoriesClient().GetPackageRepositoryDetail({ packageRepoRef });
  }

  public static async addPackageRepository(
    cluster: string,
    namespace: string,
    request: IPkgRepoFormData,
    namespaceScoped: boolean,
  ) {
    const addPackageRepositoryRequest = PackageRepositoriesService.buildAddOrUpdateRequest(
      false,
      cluster,
      namespace,
      request,
      namespaceScoped,
<<<<<<< HEAD
      this.buildEncodedCustomDetail(request),
=======
      PackageRepositoriesService.buildEncodedCustomDetail(request),
>>>>>>> 4764-pkg-repos-ui
    );

    return await this.coreRepositoriesClient().AddPackageRepository(addPackageRepositoryRequest);
  }

  public static async updatePackageRepository(
    cluster: string,
    namespace: string,
    request: IPkgRepoFormData,
  ) {
    const updatePackageRepositoryRequest = PackageRepositoriesService.buildAddOrUpdateRequest(
      true,
      cluster,
      namespace,
      request,
      undefined,
<<<<<<< HEAD
      this.buildEncodedCustomDetail(request),
=======
      PackageRepositoriesService.buildEncodedCustomDetail(request),
>>>>>>> 4764-pkg-repos-ui
    );

    return await this.coreRepositoriesClient().UpdatePackageRepository(
      updatePackageRepositoryRequest,
    );
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
    pluginCustomDetail?: any,
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
<<<<<<< HEAD
    // if using the Helm plugin, add its custom fields.
    // An "Any" object has  "typeUrl" with the FQN of the type and a "value",
=======
    // if using a plugin with customDetail, encode its custom fields,
    // otherwise skip it
    if (!request.customDetail) {
      return;
    }
    // An "Any" object has "typeUrl" with the FQN of the type and a "value",
>>>>>>> 4764-pkg-repos-ui
    // which is the result of the encoding (+finish(), to get the Uint8Array)
    // of the actual custom object
    switch (request.plugin?.name) {
      case PluginNames.PACKAGES_HELM:
        // populate the non-optional fields
        // eslint-disable-next-line no-case-declarations
        const helmCustomDetail: HelmPackageRepositoryCustomDetail = {
          ociRepositories: request.customDetail.ociRepositories || [],
          performValidation: !!request.customDetail.performValidation,
        };
        // populate the filterRule if it's not empty
        if (request.customDetail.filterRule) {
          helmCustomDetail.filterRule = request.customDetail.filterRule;
        }
        // populate the imagesPullSecret if it's not empty
        if (
          request.customDetail.imagesPullSecret?.secretRef ||
          Object.values(request.customDetail?.imagesPullSecret?.credentials as any).some(e => !!e)
        ) {
          helmCustomDetail.imagesPullSecret = request.customDetail.imagesPullSecret;
        }
        return {
          typeUrl: `${helmProtobufPackage}.HelmPackageRepositoryCustomDetail`,
<<<<<<< HEAD
          value: HelmPackageRepositoryCustomDetail.encode(helmCustomDetail).finish(),
=======
          value: HelmPackageRepositoryCustomDetail.encode(
            request.customDetail as HelmPackageRepositoryCustomDetail,
          ).finish(),
        } as Any;
      case PluginNames.PACKAGES_KAPP:
        return {
          typeUrl: `${kappControllerProtobufPackage}.KappControllerPackageRepositoryCustomDetail`,
          value: KappControllerPackageRepositoryCustomDetail.encode(
            request.customDetail as KappControllerPackageRepositoryCustomDetail,
          ).finish(),
>>>>>>> 4764-pkg-repos-ui
        } as Any;
      default:
        return;
    }
  }
}
