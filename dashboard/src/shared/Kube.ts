import * as url from "shared/url";
import { Auth } from "./Auth";
import { axiosWithAuth } from "./AxiosInstance";
import { IK8sList, IKubeState, IResource } from "./types";

export const APIBase = (cluster: string) => `api/clusters/${cluster}`;
export let WebSocketAPIBase: string;
if (window.location.protocol === "https:") {
  WebSocketAPIBase = `wss://${window.location.host}${window.location.pathname}`;
} else {
  WebSocketAPIBase = `ws://${window.location.host}${window.location.pathname}`;
}

// Kube is a lower-level class for interacting with the Kubernetes API. Use
// ResourceRef to interact with a single API resource rather than using Kube
// directly.
export class Kube {
  public static getResourceURL(
    cluster: string,
    apiVersion: string,
    resource: string,
    namespaced: boolean,
    namespace?: string,
    name?: string,
    query?: string,
  ) {
    let u = `${url.api.k8s.base(cluster)}/${
      apiVersion === "v1" || !apiVersion ? "api/v1" : `apis/${apiVersion}`
    }`;
    if (namespaced && namespace) {
      u += `/namespaces/${namespace}`;
    }
    u += `/${resource}`;
    if (name) {
      u += `/${name}`;
    }
    if (query) {
      u += `?${query}`;
    }
    return u;
  }

  public static watchResourceURL(
    cluster: string,
    apiVersion: string,
    resource: string,
    namespaced: boolean,
    namespace?: string,
    name?: string,
    query?: string,
  ) {
    let u = this.getResourceURL(cluster, apiVersion, resource, namespaced, namespace);
    u = `${WebSocketAPIBase}${u}?watch=true`;
    if (name) {
      u += `&fieldSelector=metadata.name%3D${name}`;
    }
    if (query) {
      u += `&${query}`;
    }
    return u;
  }

  public static async getResource(
    cluster: string,
    apiVersion: string,
    resource: string,
    namespaced: boolean,
    namespace?: string,
    name?: string,
    query?: string,
  ) {
    const { data } = await axiosWithAuth.get<IResource | IK8sList<IResource, {}>>(
      this.getResourceURL(cluster, apiVersion, resource, namespaced, namespace, name, query),
    );
    return data;
  }

  // Opens and returns a WebSocket for the requested resource. Note: it is
  // important that this socket be properly closed when no longer needed. The
  // returned WebSocket can be attached to an event listener to read data from
  // the socket.
  public static watchResource(
    cluster: string,
    apiVersion: string,
    resource: string,
    namespaced: boolean,
    namespace?: string,
    name?: string,
    query?: string,
  ) {
    return new WebSocket(
      this.watchResourceURL(cluster, apiVersion, resource, namespaced, namespace, name, query),
      Auth.wsProtocols(),
    );
  }

  public static async getAPIGroups(cluster: string) {
    const { data: apiGroups } = await axiosWithAuth.get(`${url.api.k8s.base(cluster)}/apis`);
    return apiGroups.groups;
  }

  public static async getResourceKinds(cluster: string, groups: any[]) {
    const result: IKubeState["kinds"] = {};
    const addResource = (r: any, version: string) => {
      // Exclude subresources
      if (!r.name.includes("/")) {
        result[r.kind] = {
          apiVersion: version,
          plural: r.name,
          namespaced: r.namespaced,
        };
      }
    };
    // Handle v1 separately
    const { data: coreResourceList } = await axiosWithAuth.get(
      `${url.api.k8s.base(cluster)}/api/v1`,
    );
    coreResourceList.resources?.forEach((r: any) => addResource(r, "v1"));

    await Promise.all(
      groups.map(async (g: any) => {
        const groupVersion = g.preferredVersion.groupVersion;
        const { data: resourceList } = await axiosWithAuth.get(
          `${url.api.k8s.base(cluster)}/apis/${groupVersion}`,
        );
        resourceList.resources?.forEach((r: any) => addResource(r, groupVersion));
      }),
    );
    return result;
  }

  public static async canI(
    cluster: string,
    group: string,
    resource: string,
    verb: string,
    namespace: string,
  ) {
    try {
      const { data } = await axiosWithAuth.post<{ allowed: boolean }>(url.backend.canI(cluster), {
        group,
        resource,
        verb,
        namespace,
      });
      return data?.allowed ? data.allowed : false;
    } catch (err) {
      return false;
    }
  }
}
