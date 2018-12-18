import { axios } from "./Auth";
import { IResource } from "./types";

export const KUBE_ROOT_URL = "api/kube";

export class Kube {
  public static getResourceURL(
    apiVersion: string,
    resource: string,
    namespace?: string,
    name?: string,
    query?: string,
  ) {
    let url = `${KUBE_ROOT_URL}/${apiVersion === "v1" ? "api/v1" : `apis/${apiVersion}`}`;
    if (namespace) {
      url += `/namespaces/${namespace}`;
    }
    url += `/${resource}`;
    if (name) {
      url += `/${name}`;
    }
    if (query) {
      url += `?${query}`;
    }
    return url;
  }

  public static async getResource(
    apiVersion: string,
    resource: string,
    namespace?: string,
    name?: string,
    query?: string,
  ) {
    const { data } = await axios.get<IResource>(
      this.getResourceURL(apiVersion, resource, namespace, name, query),
    );
    return data;
  }
}
