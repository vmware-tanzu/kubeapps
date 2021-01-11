import { filter, matches } from "lodash";
import { Kube } from "./Kube";
import { IClusterServiceVersionCRDResource, IK8sList, IKind, IResource } from "./types";

export function fromCRD(
  r: IClusterServiceVersionCRDResource,
  kind: IKind,
  cluster: string,
  namespace: string,
  ownerReference: any,
) {
  const resource = {
    apiVersion: kind.apiVersion,
    kind: r.kind,
    metadata: {},
  } as IResource;
  const ref = new ResourceRef(resource, cluster, kind.plural, kind.namespaced, namespace);
  ref.filter = {
    metadata: { ownerReferences: [ownerReference] },
  };
  return ref;
}

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
    r: IResource,
    cluster: string,
    plural: string,
    namespaced: boolean,
    defaultNamespace?: string,
  ) {
    this.cluster = cluster;
    this.plural = plural;
    this.apiVersion = r.apiVersion;
    this.kind = r.kind;
    this.name = r.metadata.name;
    this.namespace = namespaced ? r.metadata.namespace || defaultNamespace || "" : "";
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
