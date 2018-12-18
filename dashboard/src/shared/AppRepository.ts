import { axios } from "./Auth";
import { IAppRepository, IAppRepositoryList } from "./types";

export class AppRepository {
  public static async list(namespace: string) {
    const { data } = await axios.get<IAppRepositoryList>(AppRepository.getResourceLink(namespace));
    return data;
  }

  public static async get(name: string, namespace: string) {
    const { data } = await axios.get(AppRepository.getSelfLink(name, namespace));
    return data;
  }

  public static async update(name: string, namespace: string, newApp: IAppRepository) {
    const { data } = await axios.put(AppRepository.getSelfLink(name, namespace), newApp);
    return data;
  }

  public static async delete(name: string, namespace: string) {
    const { data } = await axios.delete(AppRepository.getSelfLink(name, namespace));
    return data;
  }

  public static async create(name: string, namespace: string, url: string, auth: any) {
    const { data } = await axios.post<IAppRepository>(AppRepository.getResourceLink(namespace), {
      apiVersion: "kubeapps.com/v1alpha1",
      kind: "AppRepository",
      metadata: {
        name,
      },
      spec: { auth, type: "helm", url },
    });
    return data;
  }

  private static APIBase: string = "api/kube";
  private static APIEndpoint: string = `${AppRepository.APIBase}/apis/kubeapps.com/v1alpha1`;
  private static getResourceLink(namespace?: string): string {
    return `${AppRepository.APIEndpoint}/${
      namespace ? `namespaces/${namespace}/` : ""
    }apprepositories`;
  }
  private static getSelfLink(name: string, namespace: string): string {
    return `${AppRepository.APIEndpoint}/namespaces/${namespace}/apprepositories/${name}`;
  }
}
