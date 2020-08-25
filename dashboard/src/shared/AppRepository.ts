import { axiosWithAuth } from "./AxiosInstance";
import { APIBase } from "./Kube";
import { ICreateAppRepositoryResponse } from "./types";
import * as url from "./url";

export class AppRepository {
  public static async list(namespace: string) {
    const { data } = await axiosWithAuth.get(AppRepository.getSelfLink(namespace));
    return data;
  }

  public static async get(name: string, namespace: string) {
    const { data } = await axiosWithAuth.get(AppRepository.getSelfLink(namespace, name));
    return data;
  }

  public static async resync(name: string, namespace: string) {
    const repo = await AppRepository.get(name, namespace);
    repo.spec.resyncRequests = repo.spec.resyncRequests || 0;
    repo.spec.resyncRequests++;
    const { data } = await axiosWithAuth.put(AppRepository.getSelfLink(namespace, name), repo);
    return data;
  }

  public static async update(
    name: string,
    namespace: string,
    repoURL: string,
    authHeader: string,
    customCA: string,
    syncJobPodTemplate: any,
    registrySecrets: string[],
  ) {
    const { data } = await axiosWithAuth.put<ICreateAppRepositoryResponse>(
      url.backend.apprepositories.update(namespace, name),
      {
        appRepository: { name, repoURL, authHeader, customCA, syncJobPodTemplate, registrySecrets },
      },
    );
    return data;
  }

  public static async delete(name: string, namespace: string) {
    const { data } = await axiosWithAuth.delete(
      url.backend.apprepositories.delete(name, namespace),
    );
    return data;
  }

  // create uses the kubeapps backend API
  // TODO(mnelson) Update other endpoints to similarly use the backend API, removing the need
  // for direct k8s api access (for this resource, at least).
  public static async create(
    name: string,
    namespace: string,
    repoURL: string,
    authHeader: string,
    customCA: string,
    syncJobPodTemplate: any,
    registrySecrets: string[],
  ) {
    const { data } = await axiosWithAuth.post<ICreateAppRepositoryResponse>(
      url.backend.apprepositories.create(namespace),
      {
        appRepository: { name, repoURL, authHeader, customCA, syncJobPodTemplate, registrySecrets },
      },
    );
    return data;
  }

  public static async validate(repoURL: string, authHeader: string, customCA: string) {
    const { data } = await axiosWithAuth.post<any>(url.backend.apprepositories.validate(), {
      appRepository: { repoURL, authHeader, customCA },
    });
    return data;
  }

  private static APIBase: string = APIBase;
  private static APIEndpoint: string = `${AppRepository.APIBase}/apis/kubeapps.com/v1alpha1`;
  private static getSelfLink(namespace: string, name?: string): string {
    return `${AppRepository.APIEndpoint}/namespaces/${namespace}/apprepositories${
      name ? `/${name}` : ""
    }`;
  }
}
