import { axiosWithAuth } from "./AxiosInstance";
import * as url from "./url";

import { ForbiddenError, IResource, NotFoundError } from "./types";

export default class Namespace {
  public static async list(cluster: string) {
    // This call is hitting an actual backend endpoint (see pkg/http-handler.go)
    // while the other calls (create, get) are hitting the k8s API via the
    // frountend nginx.
    const { data } = await axiosWithAuth.get<{ namespaces: IResource[] }>(
      url.backend.namespaces.list(cluster),
    );
    return data;
  }

  public static async create(cluster: string, name: string) {
    const { data } = await axiosWithAuth.post<IResource>(url.api.k8s.namespaces(cluster), {
      apiVersion: "v1",
      kind: "Namespace",
      metadata: {
        name,
      },
    });
    return data;
  }

  public static async get(cluster: string, name: string) {
    try {
      const { data } = await axiosWithAuth.get<IResource>(url.api.k8s.namespace(cluster, name));
      return data;
    } catch (err) {
      switch (err.constructor) {
        case ForbiddenError:
          throw new ForbiddenError(
            `You don't have sufficient permissions to use the namespace ${name}`,
          );
        case NotFoundError:
          throw new NotFoundError(`Namespace ${name} not found. Create it before using it.`);
        default:
          throw err;
      }
    }
  }
}

// Set of namespaces used accross the applications as default and "all ns" placeholders
export const definedNamespaces = {
  all: "_all",
};
