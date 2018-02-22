import axios from "axios";

import { IAppRepository, IAppRepositoryList } from "./types";

export class AppRepository {
  public static async list() {
    const { data } = await axios.get<IAppRepositoryList>(AppRepository.APIEndpoint);
    return data;
  }

  public static async get(name: string) {
    const { data } = await axios.get(AppRepository.getSelfLink(name));
    return data;
  }

  public static async update(name: string, newApp: IAppRepository) {
    const { data } = await axios.put(AppRepository.getSelfLink(name), newApp);
    return data;
  }

  public static async delete(name: string) {
    const { data } = await axios.delete(AppRepository.getSelfLink(name));
    return data;
  }

  public static async create(name: string, url: string) {
    const { data } = await axios.post<IAppRepository>(AppRepository.APIEndpoint, {
      apiVersion: "kubeapps.com/v1alpha1",
      kind: "AppRepository",
      metadata: {
        name,
      },
      spec: { type: "helm", url },
    });
    return data;
  }

  private static APIEndpoint: string = "/api/kube/apis/kubeapps.com/v1alpha1/namespaces/kubeapps/apprepositories";
  private static getSelfLink(name: string): string {
    return `${AppRepository.APIEndpoint}/${name}`;
  }
}
