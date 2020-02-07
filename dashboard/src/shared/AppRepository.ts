import { axiosWithAuth } from "./AxiosInstance";
import { APIBase } from "./Kube";
import { definedNamespaces } from "./Namespace";
import { IAppRepository, IAppRepositoryList, ICreateAppRepositoryResponse } from "./types";
import * as url from "./url";

export class AppRepository {
  public static async list(namespace: string) {
    const { data } = await axiosWithAuth.get<IAppRepositoryList>(
      AppRepository.getResourceLink(namespace),
    );
    return data;
  }

  public static async get(name: string, namespace: string) {
    const { data } = await axiosWithAuth.get(AppRepository.getSelfLink(name, namespace));
    return data;
  }

  public static async update(name: string, namespace: string, newApp: IAppRepository) {
    const { data } = await axiosWithAuth.put(AppRepository.getSelfLink(name, namespace), newApp);
    return data;
  }

  public static async delete(name: string, namespace: string) {
    const { data } = await axiosWithAuth.delete(AppRepository.getSelfLink(name, namespace));
    return data;
  }

  // create uses the kubeapps backend API
  // TODO(mnelson) Update other endpoints to similarly use the backend API, removing the need
  // for direct k8s api access (for this resource, at least).
  public static async create(
    name: string,
    repoURL: string,
    authHeader: string,
    customCA: string,
    syncJobPodTemplate: any,
  ) {
    const { data } = await axiosWithAuth.post<ICreateAppRepositoryResponse>(
      url.backend.apprepositories.create(),
      { appRepository: { name, repoURL, authHeader, customCA, syncJobPodTemplate } },
    );
    return data;
  }

  private static APIBase: string = APIBase;
  private static APIEndpoint: string = `${AppRepository.APIBase}/apis/kubeapps.com/v1alpha1`;
  private static getResourceLink(namespace?: string): string {
    return `${AppRepository.APIEndpoint}/${
      !namespace || namespace === definedNamespaces.all ? "" : `namespaces/${namespace}/`
    }apprepositories`;
  }
  private static getSelfLink(name: string, namespace: string): string {
    return `${AppRepository.APIEndpoint}/namespaces/${namespace}/apprepositories/${name}`;
  }
}
