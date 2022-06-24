// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { axiosWithAuth } from "./AxiosInstance";
import { ICreateAppRepositoryResponse, IPkgRepoFormData } from "./types";
import * as url from "./url";

export class AppRepository {
  public static async list(cluster: string, namespace: string) {
    const {
      data: { appRepository },
    } = await axiosWithAuth.get<any>(url.backend.apprepositories.list(cluster, namespace));
    return appRepository;
  }

  public static async get(cluster: string, namespace: string, name: string) {
    const {
      data: { appRepository },
    } = await axiosWithAuth.get<any>(url.backend.apprepositories.get(cluster, namespace, name));
    return appRepository;
  }

  public static async update(cluster: string, namespace: string, request: IPkgRepoFormData) {
    const { data } = await axiosWithAuth.put<ICreateAppRepositoryResponse>(
      url.backend.apprepositories.update(cluster, namespace, request.name),
      {
        appRepository: {
          name: request.name,
          repoURL: request.url,
          type: request.type,
          description: request.description,
          authHeader: request.authHeader,
          authRegCreds: request.customDetails.dockerRegistrySecrets,
          customCA: request.customCA,
          syncJobPodTemplate: "",
          registrySecrets: request.dockerRegCreds,
          ociRepositories: request.customDetails.ociRepositories,
          tlsInsecureSkipVerify: request.skipTLS,
          passCredentials: request.passCredentials,
          filterRule: request.customDetails.filterRule,
        },
      },
    );
    return data;
  }

  public static async delete(cluster: string, namespace: string, name: string) {
    const { data } = await axiosWithAuth.delete(
      url.backend.apprepositories.delete(cluster, namespace, name),
    );
    return data;
  }

  // create uses the kubeapps backend API
  // TODO(mnelson) Update other endpoints to similarly use the backend API, removing the need
  // for direct k8s api access (for this resource, at least).
  public static async create(cluster: string, namespace: string, request: IPkgRepoFormData) {
    const { data } = await axiosWithAuth.post<ICreateAppRepositoryResponse>(
      url.backend.apprepositories.create(cluster, namespace),
      {
        appRepository: {
          name: request.name,
          repoURL: request.url,
          type: request.type,
          description: request.description,
          authHeader: request.authHeader,
          authRegCreds: request.customDetails.dockerRegistrySecrets,
          customCA: request.customCA,
          syncJobPodTemplate: "",
          registrySecrets: request.dockerRegCreds,
          ociRepositories: request.customDetails.ociRepositories,
          tlsInsecureSkipVerify: request.skipTLS,
          passCredentials: request.passCredentials,
          filterRule: request.customDetails.filterRule,
        },
      },
    );
    return data;
  }

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
}
