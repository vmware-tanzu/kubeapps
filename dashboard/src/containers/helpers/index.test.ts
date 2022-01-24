// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { keyForResourceRef } from "shared/ResourceRef";
import { IKubeItem, IKubeState, IResource } from "shared/types";
import { filterByResourceRefs } from ".";
import { ResourceRef } from "gen/kubeappsapis/core/packages/v1alpha1/packages";

describe("filterByResourceRefs", () => {
  const svc1 = {
    apiVersion: "v1",
    kind: "Service",
    metadata: { name: "bar", namespace: "foo" },
  } as IResource;
  const svc1Ref = {
    apiVersion: "v1",
    kind: "Service",
    name: "bar",
    namespace: "foo",
  } as ResourceRef;
  const svc1Key = keyForResourceRef(svc1Ref);
  const svc2 = {
    apiVersion: "v1",
    kind: "Service",
    metadata: { name: "bar", namespace: "foo1" },
  } as IResource;
  const svc2Ref = {
    apiVersion: "v1",
    kind: "Service",
    name: "bar",
    namespace: "foo1",
  } as ResourceRef;
  const svc2Key = keyForResourceRef(svc2Ref);
  const deploy = {
    apiVersion: "apps/v1",
    kind: "Deployment",
    metadata: { name: "bar", namespace: "foo1" },
  } as IResource;

  const items: IKubeState["items"] = {
    [svc1Key]: {
      item: svc1,
    } as IKubeItem<IResource>,
    [svc2Key]: {
      item: svc2,
    } as IKubeItem<IResource>,
    "unused-key": {
      item: deploy,
    } as IKubeItem<IResource>,
  };
  it("returns the IKubeItems in the state referenced by each ResourceRef", () => {
    const resourceRefs: ResourceRef[] = [svc1Ref, svc2Ref];

    expect(filterByResourceRefs(resourceRefs, items)).toEqual([{ item: svc1 }, { item: svc2 }]);
  });

  it("does not return resources that are not in the state", () => {
    const missingSvcRef = {
      apiVersion: "v1",
      kind: "Service",
      name: "missing",
      namespace: "foo1",
    } as ResourceRef;
    const resourceRefs: ResourceRef[] = [svc2Ref, missingSvcRef];

    expect(filterByResourceRefs(resourceRefs, items)).toEqual([{ item: svc2 }]);
  });
});
