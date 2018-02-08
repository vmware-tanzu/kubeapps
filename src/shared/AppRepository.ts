import axios from "axios";

import { IAppRepository, IAppRepositoryList } from "./types";

export class AppRepository {
  public static async list() {
    const { data } = await axios.get<IAppRepositoryList>(
      `${AppRepository.APIEndpoint}/apprepositories`,
    );
    return data;
  }

  public static async delete(name: string, namespace: string = "default") {
    const { data } = await axios.delete(AppRepository.getSelfLink(name, namespace));
    return data;
  }

  public static async create(name: string, url: string, namespace: string = "default") {
    const { data } = await axios.post<IAppRepository>(
      `${AppRepository.APIEndpoint}/namespaces/${namespace}/apprepositories`,
      {
        apiVersion: "kubeapps.com/v1alpha1",
        kind: "AppRepository",
        metadata: {
          name,
          namespace,
        },
        spec: { type: "helm", url },
      },
    );
    return data;
  }

  // private static serviceCatalogURL: string = "https://svc-catalog-charts.storage.googleapis.com";
  private static APIEndpoint: string = "/api/kube/apis/kubeapps.com/v1alpha1";
  private static getSelfLink(name: string, namespace: string = "default"): string {
    return `${AppRepository.APIEndpoint}/namespaces/${namespace}/apprepositories/${name}`;
  }
}
