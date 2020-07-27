import { filter, matches } from "lodash";
import { Kube } from "./Kube";
import { ResourceKind } from "./ResourceKinds";
import { IClusterServiceVersionCRDResource, IK8sList, IResource } from "./types";

export function fromCRD(
  r: IClusterServiceVersionCRDResource,
  cluster: string,
  namespace: string,
  ownerReference: any,
) {
  const resource = {
    apiVersion: Kube.resourceAPIVersion(r.kind as ResourceKind),
    kind: r.kind,
    metadata: {},
  } as IResource;
  // Avoid namespace for cluster-wide resources supported (ClusterRole, ClusterRoleBinding)
  // TODO(andresmgot): This won't work for new resource types, we would need to dinamically
  // resolve those
  const resourceNamespace = r.kind.startsWith("Cluster") ? "" : namespace;
  const ref = new ResourceRef(resource, cluster, resourceNamespace);
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
  public kind: ResourceKind;
  public name: string;
  public namespace: string;
  public filter: any;

  // Creates a new ResourceRef instance from an existing IResource. Provide
  // defaultNamespace to set if the IResource doesn't specify a namespace.
  constructor(r: IResource, cluster: string, defaultNamespace?: string) {
    this.cluster = cluster;
    this.apiVersion = r.apiVersion;
    this.kind = r.kind;
    this.name = r.metadata.name;
    this.namespace = r.metadata.namespace || defaultNamespace || "";
    return this;
  }

  // Gets a full resource URL for the referenced resource
  public getResourceURL() {
    return Kube.getResourceURL(
      this.cluster,
      this.apiVersion,
      Kube.resourcePlural(this.kind),
      this.namespace,
      this.name,
    );
  }

  public watchResourceURL() {
    return Kube.watchResourceURL(
      this.cluster,
      this.apiVersion,
      Kube.resourcePlural(this.kind),
      this.namespace,
      this.name,
    );
  }

  public async getResource() {
    const resource = await Kube.getResource(
      this.cluster,
      this.apiVersion,
      Kube.resourcePlural(this.kind),
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
      Kube.resourcePlural(this.kind),
      this.namespace,
      this.name,
    );
  }
}

export default ResourceRef;
