import * as url from "shared/url";
import { Auth } from "./Auth";
import { axiosWithAuth } from "./AxiosInstance";
import { ResourceKindsWithAPIVersions } from "./ResourceAPIVersion";
import { isNamespaced, ResourceKind, ResourceKindsWithPlurals } from "./ResourceKinds";
import { IK8sList, IResource } from "./types";

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
    namespace?: string,
    name?: string,
    query?: string,
  ) {
    let u = `${url.api.k8s.base(cluster)}/${
      apiVersion === "v1" || !apiVersion ? "api/v1" : `apis/${apiVersion}`
    }`;
    if (namespace && isNamespaced(resource)) {
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
    namespace?: string,
    name?: string,
    query?: string,
  ) {
    let u = this.getResourceURL(cluster, apiVersion, resource, namespace);
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
    namespace?: string,
    name?: string,
    query?: string,
  ) {
    const { data } = await axiosWithAuth.get<IResource | IK8sList<IResource, {}>>(
      this.getResourceURL(cluster, apiVersion, resource, namespace, name, query),
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
    namespace?: string,
    name?: string,
    query?: string,
  ) {
    return new WebSocket(
      this.watchResourceURL(cluster, apiVersion, resource, namespace, name, query),
      Auth.wsProtocols(),
    );
  }

  // Gets the plural form of the resource Kind for use in the resource path
  public static resourcePlural(kind: ResourceKind) {
    return ResourceKindsWithPlurals[kind];
  }

  // Gets the apiVersion of the resource Kind for use in the resource path
  public static resourceAPIVersion(kind: ResourceKind) {
    return ResourceKindsWithAPIVersions[kind];
  }

  public static async canI(
    cluster: string,
    group: string,
    resource: string,
    verb: string,
    namespace: string,
  ) {
    const { data } = await axiosWithAuth.post<{ allowed: boolean }>(url.backend.canI(cluster), {
      group,
      resource,
      verb,
      namespace,
    });
    return data.allowed;
  }
}
