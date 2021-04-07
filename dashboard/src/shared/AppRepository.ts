import { axiosWithAuth } from "./AxiosInstance";
import { APIBase } from "./Kube";
import { IAppRepositoryFilter, ICreateAppRepositoryResponse } from "./types";
import * as url from "./url";

export class AppRepository {
  public static async list(cluster: string, namespace: string) {
    const {
      data: { appRepository },
    } = await axiosWithAuth.get(url.backend.apprepositories.list(cluster, namespace));
    return appRepository;
  }

  public static async get(cluster: string, namespace: string, name: string) {
    const { data } = await axiosWithAuth.get(AppRepository.getSelfLink(cluster, namespace, name));
    return data;
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
    authHeader: string,
    authRegCreds: string,
    customCA: string,
    syncJobPodTemplate: any,
    registrySecrets: string[],
    ociRepositories: string[],
    skipTLS: boolean,
    filter?: IAppRepositoryFilter,
  ) {
    const { data } = await axiosWithAuth.put<ICreateAppRepositoryResponse>(
      url.backend.apprepositories.update(cluster, namespace, name),
      {
        appRepository: {
          name,
          repoURL,
          type,
          authHeader,
          authRegCreds,
          customCA,
          syncJobPodTemplate,
          registrySecrets,
          ociRepositories,
          tlsInsecureSkipVerify: skipTLS,
          filter,
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
    authHeader: string,
    authRegCreds: string,
    customCA: string,
    syncJobPodTemplate: any,
    registrySecrets: string[],
    ociRepositories: string[],
    skipTLS: boolean,
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
          customCA,
          syncJobPodTemplate,
          registrySecrets,
          ociRepositories,
          tlsInsecureSkipVerify: skipTLS,
          filterRule: filter,
        },
      },
    );
    return data;
  }

  public static async validate(
    cluster: string,
    repoURL: string,
    type: string,
    authHeader: string,
    authRegCreds: string,
    customCA: string,
    ociRepositories: string[],
    skipTLS: boolean,
  ) {
    const { data } = await axiosWithAuth.post<any>(url.backend.apprepositories.validate(cluster), {
      appRepository: {
        repoURL,
        type,
        authHeader,
        authRegCreds,
        customCA,
        ociRepositories,
        tlsInsecureSkipVerify: skipTLS,
      },
    });
    return data;
  }

  private static APIEndpoint(cluster: string): string {
    return `${APIBase(cluster)}/apis/kubeapps.com/v1alpha1`;
  }
  private static getSelfLink(cluster: string, namespace: string, name?: string): string {
    return `${AppRepository.APIEndpoint(cluster)}/namespaces/${namespace}/apprepositories${
      name ? `/${name}` : ""
    }`;
  }
}
