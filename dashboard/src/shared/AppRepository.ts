import { axios } from "./Auth";
import Config from "./Config";
import { IAppRepository, IAppRepositoryList } from "./types";

export class AppRepository {
  public static async list() {
    const { data } = await axios.get<IAppRepositoryList>(await AppRepository.getResourceLink());
    return data;
  }

  public static async get(name: string) {
    const { data } = await axios.get(await AppRepository.getResourceLink(name));
    return data;
  }

  public static async update(name: string, newApp: IAppRepository) {
    const { data } = await axios.put(await AppRepository.getResourceLink(name), newApp);
    return data;
  }

  public static async delete(name: string) {
    const { data } = await axios.delete(await AppRepository.getResourceLink(name));
    return data;
  }

  public static async create(name: string, url: string, auth: any) {
    const { data } = await axios.post<IAppRepository>(await AppRepository.getResourceLink(), {
      apiVersion: "kubeapps.com/v1alpha1",
      kind: "AppRepository",
      metadata: {
        name,
      },
      spec: { auth, type: "helm", url },
    });
    return data;
  }

  private static APIBase: string = "/api/kube";
  private static APIEndpoint: string = `${AppRepository.APIBase}/apis/kubeapps.com/v1alpha1`;
  private static async getResourceLink(name?: string) {
    const { namespace } = await Config.getConfig();
    return `${AppRepository.APIEndpoint}/namespaces/${namespace}/apprepositories${
      name ? `/${name}` : ""
    }`;
  }
}
