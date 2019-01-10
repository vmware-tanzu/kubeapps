import { IResource } from "./types";

class ResourceRef {
  // Creates a new ResourceRef instance from an existing IResource.
  // Provide defaultNamespace to set if the IResource doesn't specify a namespace.
  public static newFromResource(r: IResource, defaultNamespace: string = "default") {
    const ref = new ResourceRef();
    ref.apiVersion = r.apiVersion;
    ref.kind = r.kind;
    ref.name = r.metadata.name;
    ref.namespace = r.metadata.namespace || defaultNamespace;
    return ref;
  }

  public apiVersion: string;
  public kind: string;
  public name: string;
  public namespace: string;
}

export default ResourceRef;
