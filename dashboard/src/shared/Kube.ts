import { Auth } from "./Auth";
import { axiosWithAuth } from "./AxiosInstance";
import { IResource } from "./types";

export const APIBase = "api/kube";
export let WebSocketAPIBase: string;
if (location.protocol === "https:") {
  WebSocketAPIBase = `wss://${window.location.host}${window.location.pathname}`;
} else {
  WebSocketAPIBase = `ws://${window.location.host}${window.location.pathname}`;
}

// We explicitly define the plurals here, just in case a generic pluralizer
// isn't sufficient. Note that CRDs can explicitly define pluralized forms,
// which might not match with the Kind. If this becomes difficult to
// maintain we can add a generic pluralizer and a way to override.
const ResourceKindToPlural = {
  Secret: "secrets",
  Service: "services",
  Ingress: "ingresses",
  Deployment: "deployments",
  StatefulSet: "statefulsets",
  DaemonSet: "daemonsets",
};

// Kube is a lower-level class for interacting with the Kubernetes API. Use
// ResourceRef to interact with a single API resource rather than using Kube
// directly.
export class Kube {
  public static getResourceURL(
    apiVersion: string,
    resource: string,
    namespace?: string,
    name?: string,
    query?: string,
  ) {
    let url = `${APIBase}/${apiVersion === "v1" ? "api/v1" : `apis/${apiVersion}`}`;
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

  public static watchResourceURL(
    apiVersion: string,
    resource: string,
    namespace?: string,
    name?: string,
    query?: string,
  ) {
    let url = this.getResourceURL(apiVersion, resource, namespace);
    url = `${WebSocketAPIBase}${url}?watch=true`;
    if (name) {
      url += `&fieldSelector=metadata.name%3D${name}`;
    }
    if (query) {
      url += `&${query}`;
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
    const { data } = await axiosWithAuth.get<IResource>(
      this.getResourceURL(apiVersion, resource, namespace, name, query),
    );
    return data;
  }

  // Opens and returns a WebSocket for the requested resource. Note: it is
  // important that this socket be properly closed when no longer needed. The
  // returned WebSocket can be attached to an event listener to read data from
  // the socket.
  public static watchResource(
    apiVersion: string,
    resource: string,
    namespace?: string,
    name?: string,
    query?: string,
  ) {
    return new WebSocket(
      this.watchResourceURL(apiVersion, resource, namespace, name, query),
      Auth.wsProtocols(),
    );
  }

  // Gets the plural form of the resource Kind for use in the resource path
  public static resourcePlural(kind: string) {
    const plural = ResourceKindToPlural[kind];
    if (!plural) {
      throw new Error(`Don't know plural for ${kind}, register it in Kube`);
    }
    return plural;
  }
}
