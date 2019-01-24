import { filterByResourceRefs } from ".";
import ResourceRef from "../../shared/ResourceRef";
import { IKubeItem, IKubeState, IResource } from "../../shared/types";

describe("filterByResourceRefs", () => {
  const svc1 = {
    apiVersion: "v1",
    kind: "Service",
    metadata: { name: "bar", namespace: "foo" },
  } as IResource;
  const svc2 = {
    apiVersion: "v1",
    kind: "Service",
    metadata: { name: "bar", namespace: "foo1" },
  } as IResource;
  const deploy = {
    apiVersion: "apps/v1",
    kind: "Deployment",
    metadata: { name: "bar", namespace: "foo1" },
  } as IResource;

  const items: IKubeState["items"] = {
    "api/kube/api/v1/namespaces/foo/services/bar": { item: svc1 } as IKubeItem<IResource>,
    "api/kube/api/v1/namespaces/foo1/services/bar": { item: svc2 } as IKubeItem<IResource>,
    "api/kube/apis/apps/v1/namespaces/foo1/deployments/bar": { item: deploy } as IKubeItem<
      IResource
    >,
  };
  it("returns the IKubeItems in the state referenced by each ResourceRef", () => {
    const resourceRefs: ResourceRef[] = [new ResourceRef(svc1), new ResourceRef(svc2)];

    expect(filterByResourceRefs(resourceRefs, items)).toEqual([{ item: svc1 }, { item: svc2 }]);
  });

  it("does not return resources that are not in the state", () => {
    const missingSvc = {
      apiVersion: "v1",
      kind: "Service",
      metadata: {
        name: "missing",
        namespace: "foo1",
      },
    } as IResource;
    const resourceRefs: ResourceRef[] = [new ResourceRef(svc2), new ResourceRef(missingSvc)];

    expect(filterByResourceRefs(resourceRefs, items)).toEqual([{ item: svc2 }]);
  });
});
