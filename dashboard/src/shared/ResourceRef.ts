import { Kube } from "./Kube";
import { IResource } from "./types";

// ResourceRef defines a reference to a namespaced Kubernetes API Object and
// provides helpers to retrieve the resource URL
class ResourceRef {
  public apiVersion: string;
  public kind: string;
  public name: string;
  public namespace: string;

  // Creates a new ResourceRef instance from an existing IResource. Provide
  // defaultNamespace to set if the IResource doesn't specify a namespace.
  // TODO: add support for cluster-scoped resources, or add a ClusterResourceRef
  // class.
  constructor(r: IResource, defaultNamespace?: string) {
    this.apiVersion = r.apiVersion;
    this.kind = r.kind;
    this.name = r.metadata.name;
    const namespace = r.metadata.namespace || defaultNamespace;
    if (!namespace) {
      throw new Error(`Namespace missing for resource ${this.name}, define a default namespace`);
    }
    this.namespace = namespace;
    return this;
  }

  // Gets a full resource URL for the referenced resource
  public getResourceURL() {
    return Kube.getResourceURL(
      this.apiVersion,
      Kube.resourcePlural(this.kind),
      this.namespace,
      this.name,
    );
  }

  public watchResourceURL() {
    return Kube.watchResourceURL(
      this.apiVersion,
      Kube.resourcePlural(this.kind),
      this.namespace,
      this.name,
    );
  }

  public getResource() {
    return Kube.getResource(
      this.apiVersion,
      Kube.resourcePlural(this.kind),
      this.namespace,
      this.name,
    );
  }

  // Opens and returns a WebSocket for the requested resource. Note: it is
  // important that this socket be properly closed when no longer needed. The
  // returned WebSocket can be attached to an event listener to read data from
  // the socket.
  public watchResource() {
    return Kube.watchResource(
      this.apiVersion,
      Kube.resourcePlural(this.kind),
      this.namespace,
      this.name,
    );
  }
}

export default ResourceRef;
