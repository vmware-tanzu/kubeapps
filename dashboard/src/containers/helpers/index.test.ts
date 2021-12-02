import ResourceRef, { keyForResourceRef } from "shared/ResourceRef";
import { IKubeItem, IKubeState, IResource } from "shared/types";
import { filterByResourceRefs } from ".";
import { ResourceRef as APIResourceRef } from "gen/kubeappsapis/core/packages/v1alpha1/packages";

const clusterName = "cluster-name";

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
  } as APIResourceRef;
  const svc1Key = keyForResourceRef(
    svc1Ref.apiVersion,
    svc1Ref.kind,
    svc1Ref.namespace,
    svc1Ref.name,
  );
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
  } as APIResourceRef;
  const svc2Key = keyForResourceRef(
    svc2Ref.apiVersion,
    svc2Ref.kind,
    svc2Ref.namespace,
    svc2Ref.name,
  );
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
    const resourceRefs: ResourceRef[] = [
      new ResourceRef(svc1Ref, clusterName, "services", true, "foo"),
      new ResourceRef(svc2Ref, clusterName, "services", true, "foo1"),
    ];

    expect(filterByResourceRefs(resourceRefs, items)).toEqual([{ item: svc1 }, { item: svc2 }]);
  });

  it("does not return resources that are not in the state", () => {
    const missingSvcRef = {
      apiVersion: "v1",
      kind: "Service",
      name: "missing",
      namespace: "foo1",
    } as APIResourceRef;
    const resourceRefs: ResourceRef[] = [
      new ResourceRef(svc2Ref, clusterName, "services", true, "default"),
      new ResourceRef(missingSvcRef, clusterName, "services", true, "default"),
    ];

    expect(filterByResourceRefs(resourceRefs, items)).toEqual([{ item: svc2 }]);
  });
});
