import { filter, matches } from "lodash";
import { Kube } from "./Kube";
import { IClusterServiceVersionCRDResource, IK8sList, IKind, IResource } from "./types";
import { ResourceRef as APIResourceRef } from "gen/kubeappsapis/core/packages/v1alpha1/packages";

export function fromCRD(
  r: IClusterServiceVersionCRDResource,
  kind: IKind,
  cluster: string,
  namespace: string,
  ownerReference: any,
) {
  const apiResourceRef = {
    apiVersion: kind.apiVersion,
    kind: r.kind,
  } as APIResourceRef;
  const ref = new ResourceRef(apiResourceRef, cluster, kind.plural, kind.namespaced, namespace);
  ref.filter = {
    metadata: { ownerReferences: [ownerReference] },
  };
  return ref;
}

// TODO(minelson): Update to use API resourceRef type once old model removed.
export const keyForResourceRef = (
  apiVersion: string,
  kind: string,
  namespace: string,
  name: string,
) => `${apiVersion}/${kind}/${namespace}/${name}`;

// ResourceRef defines a reference to a namespaced Kubernetes API Object and
// provides helpers to retrieve the resource URL
class ResourceRef {
  public cluster: string;
  public apiVersion: string;
  public kind: string;
  public plural: string;
  public namespaced: boolean;
  public name: string;
  public namespace: string;
  public filter: any;

  // Creates a new ResourceRef instance from an existing IResource. Provide
  // defaultNamespace to set if the IResource doesn't specify a namespace.
  constructor(
    apiRef: APIResourceRef,
    cluster: string,
    plural: string,
    namespaced: boolean,
    releaseNamespace: string,
  ) {
    this.cluster = cluster;
    this.plural = plural;
    this.apiVersion = apiRef.apiVersion;
    this.kind = apiRef.kind;
    this.name = apiRef.name;
    this.namespace = namespaced ? apiRef.namespace || releaseNamespace || "" : "";
    this.namespaced = namespaced;
    return this;
  }

  // Gets a full resource URL for the referenced resource
  public getResourceURL() {
    return Kube.getResourceURL(
      this.cluster,
      this.apiVersion,
      this.plural,
      this.namespaced,
      this.namespace,
      this.name,
    );
  }

  public watchResourceURL() {
    return Kube.watchResourceURL(
      this.cluster,
      this.apiVersion,
      this.plural,
      this.namespaced,
      this.namespace,
      this.name,
    );
  }

  public async getResource() {
    const resource = await Kube.getResource(
      this.cluster,
      this.apiVersion,
      this.plural,
      this.namespaced,
      this.namespace,
      this.name,
    );
    const resourceList = resource as IK8sList<IResource, {}>;
    if (resourceList.items) {
      resourceList.items = filter(resourceList.items, matches(this.filter));
      return resourceList;
    }
    return resource;
  }

  // Opens and returns a WebSocket for the requested resource. Note: it is
  // important that this socket be properly closed when no longer needed. The
  // returned WebSocket can be attached to an event listener to read data from
  // the socket.
  public watchResource() {
    return Kube.watchResource(
      this.cluster,
      this.apiVersion,
      this.plural,
      this.namespaced,
      this.namespace,
      this.name,
    );
  }
}

export default ResourceRef;
