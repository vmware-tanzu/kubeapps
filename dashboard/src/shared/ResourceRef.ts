// Copyright 2019-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import type { PartialMessage } from "@bufbuild/protobuf";
import { IClusterServiceVersionCRDResource, IKind } from "./types";
import { ResourceRef as APIResourceRef } from "gen/kubeappsapis/core/packages/v1alpha1/packages_pb";

export function fromCRD(
  r: IClusterServiceVersionCRDResource,
  kind: IKind,
  cluster: string,
  namespace: string,
  ownerReference: any,
) {
  const ref = new ResourceRef(cluster, kind.plural, kind.namespaced, namespace, {
    apiVersion: kind.apiVersion,
    kind: r.kind,
  });
  ref.filter = {
    metadata: { ownerReferences: [ownerReference] },
  };
  return ref;
}

// keyForResourceRef is used to create a key for the redux state tracking resources
// keyed by references.
export const keyForResourceRef = (r: APIResourceRef) =>
  `${r.apiVersion}/${r.kind}/${r.namespace}/${r.name}`;

// ResourceRef defines a reference to a namespaced Kubernetes API Object and
// provides helpers to retrieve the resource URL
class ResourceRef extends APIResourceRef {
  public cluster: string;
  public plural: string;
  public namespaced: boolean;
  public filter: any;

  // Creates a new ResourceRef instance from an existing IResource. Provide
  // defaultNamespace to set if the IResource doesn't specify a namespace.
  //constructor(data?: PartialMessage<ResourceRef>) {
  constructor(
    cluster: string,
    plural: string,
    namespaced: boolean,
    releaseNamespace: string,
    data?: PartialMessage<APIResourceRef>,
  ) {
    data = data || {};
    data.namespace = namespaced ? data.namespace || releaseNamespace || "" : "";
    super(data);
    this.namespaced = namespaced;
    this.cluster = cluster;
    this.plural = plural;
    return this;
  }
}

export default ResourceRef;
