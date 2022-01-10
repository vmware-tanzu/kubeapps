import { axiosWithAuth } from "./AxiosInstance";
import { IAppRepositoryFilter, ICreateAppRepositoryResponse } from "./types";
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

  public static async update(
    cluster: string,
    name: string,
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
    const { data } = await axiosWithAuth.put<ICreateAppRepositoryResponse>(
      url.backend.apprepositories.update(cluster, namespace, name),
      {
        appRepository: {
          name,
          repoURL,
          type,
          description,
          authHeader,
          authRegCreds,
          customCA,
          syncJobPodTemplate,
          registrySecrets,
          ociRepositories,
          tlsInsecureSkipVerify: skipTLS,
          passCredentials: passCredentials,
          filterRule: filter,
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
  public static async create(
    cluster: string,
    name: string,
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
    const { data } = await axiosWithAuth.post<ICreateAppRepositoryResponse>(
      url.backend.apprepositories.create(cluster, namespace),
      {
        appRepository: {
          name,
          repoURL,
          authHeader,
          authRegCreds,
          type,
          description,
          customCA,
          syncJobPodTemplate,
          registrySecrets,
          ociRepositories,
          tlsInsecureSkipVerify: skipTLS,
          passCredentials: passCredentials,
          filterRule: filter,
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
