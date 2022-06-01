// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import {
  DeletePackageRepositoryResponse,
  GetPackageRepositoryDetailResponse,
  GetPackageRepositorySummariesResponse,
  PackageRepositoryAuth_PackageRepositoryAuthType,
  PackageRepositoryReference,
} from "gen/kubeappsapis/core/packages/v1alpha1/repositories";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import { axiosWithAuth } from "./AxiosInstance";
import KubeappsGrpcClient from "./KubeappsGrpcClient";
import { IAppRepositoryFilter } from "./types";
import * as url from "./url";

export class PackageRepositoriesService {
  public static coreRepositoriesClient = () =>
    new KubeappsGrpcClient().getRepositoriesServiceClientImpl();

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
    filter?: IAppRepositoryFilter,
  ) {
    console.warn("UNUSED authRegCreds", JSON.stringify(authRegCreds));
    console.warn("UNUSED syncJobPodTemplate", JSON.stringify(syncJobPodTemplate));
    console.warn("UNUSED registrySecrets", JSON.stringify(registrySecrets));
    console.warn("UNUSED ociRepositories", JSON.stringify(ociRepositories));
    console.warn("UNUSED passCredentials", JSON.stringify(passCredentials));
    console.warn("UNUSED filter", JSON.stringify(filter));

    return await this.coreRepositoriesClient().AddPackageRepository({
      context: { cluster, namespace },
      name,
      description,
      namespaceScoped: false,
      type,
      url: repoURL,
      interval: 3600,
      tlsConfig: {
        certAuthority: customCA,
        insecureSkipVerify: skipTLS,
        // secretRef: { key: "", name: "" },
      },
      auth: {
        header: authHeader,
        dockerCreds: {
          email: "",
          password: "",
          server: "",
          username: "",
        },
        passCredentials: false,
        // secretRef: {
        //   key: "",
        //   name: "",
        // },
        // tlsCertKey: {
        //   cert: "",
        //   key: "",
        // },
        type: PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED,
        usernamePassword: {
          password: "",
          username: "",
        },
      },
      plugin,
      customDetail: {
        typeUrl: "",
        value: undefined,
      },
    });
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
    filter?: IAppRepositoryFilter,
  ) {
    console.warn("UNUSED type", JSON.stringify(type));
    console.warn("UNUSED authRegCreds", JSON.stringify(authRegCreds));
    console.warn("UNUSED syncJobPodTemplate", JSON.stringify(syncJobPodTemplate));
    console.warn("UNUSED registrySecrets", JSON.stringify(registrySecrets));
    console.warn("UNUSED ociRepositories", JSON.stringify(ociRepositories));
    console.warn("UNUSED passCredentials", JSON.stringify(passCredentials));
    console.warn("UNUSED filter", JSON.stringify(filter));

    return await this.coreRepositoriesClient().UpdatePackageRepository({
      packageRepoRef: {
        context: { cluster, namespace },
        identifier: name,
        plugin: plugin,
      },
      description,
      url: repoURL,
      interval: 3600,
      tlsConfig: {
        certAuthority: customCA,
        insecureSkipVerify: skipTLS,
        // secretRef: { key: "", name: "" },
      },
      auth: {
        header: authHeader,
        dockerCreds: { email: "", password: "", server: "", username: "" },
        passCredentials: false,
        // secretRef: { key: "", name: "" },
        tlsCertKey: {
          cert: "",
          key: "",
        },
        type: PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED,
        usernamePassword: {
          password: "",
          username: "",
        },
      },
      customDetail: {
        typeUrl: "",
        value: undefined,
      },
    });
  }

  public static async deletePackageRepository(
    packageRepoRef: PackageRepositoryReference,
  ): Promise<DeletePackageRepositoryResponse> {
    return await this.coreRepositoriesClient().DeletePackageRepository({
      packageRepoRef,
    });
  }

  // ............................... DEPRECATED ...............................

  public static async validate(
    cluster: string,
    namespace: string,
    repoURL: string,
    type: string,
    authHeader: string,
    authRegCreds: string,
    customCA: string,
    ociRepositories: string[],
    skipTLS: boolean,
    passCredentials: boolean,
  ) {
    const { data } = await axiosWithAuth.post<any>(
      url.backend.apprepositories.validate(cluster, namespace),
      {
        appRepository: {
          repoURL,
          type,
          authHeader,
          authRegCreds,
          customCA,
          ociRepositories,
          tlsInsecureSkipVerify: skipTLS,
          passCredentials: passCredentials,
        },
      },
    );
    return data;
  }

  public static async getSecretForRepo(cluster: string, namespace: string, name: string) {
    const {
      data: { secret },
    } = await axiosWithAuth.get<any>(url.backend.apprepositories.get(cluster, namespace, name));
    return secret;
  }

  public static async resync(cluster: string, namespace: string, name: string) {
    const { data } = await axiosWithAuth.post(
      url.backend.apprepositories.refresh(cluster, namespace, name),
      null,
    );
    return data;
  }
}
