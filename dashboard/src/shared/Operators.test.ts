import { axiosWithAuth } from "./AxiosInstance";
import { Operators } from "./Operators";
import { IClusterServiceVersion, IPackageManifest, IResource } from "./types";

it("check if the OLM has been installed", async () => {
  axiosWithAuth.get = jest.fn(() => {
    return { status: 200 };
  });
  expect(await Operators.isOLMInstalled()).toBe(true);
  expect(axiosWithAuth.get).toHaveBeenCalled();
  expect((axiosWithAuth.get as jest.Mock).mock.calls[0][0]).toEqual(
    "api/kube/apis/packages.operators.coreos.com/v1/packagemanifests",
  );
});

it("OLM is not installed if the request fails", async () => {
  axiosWithAuth.get = jest.fn(() => {
    return { status: 404 };
  });
  expect(await Operators.isOLMInstalled()).toBe(false);
});

it("OLM is not installed if the request returns != 200", async () => {
  axiosWithAuth.get = jest.fn(() => {
    return { status: 404 };
  });
  expect(await Operators.isOLMInstalled()).toBe(false);
});

it("get operators", async () => {
  const operator = { metadata: { name: "foo" } } as IPackageManifest;
  const ns = "default";
  axiosWithAuth.get = jest.fn(() => {
    return { data: { items: [operator] } };
  });
  expect(await Operators.getOperators(ns)).toEqual([operator]);
  expect(axiosWithAuth.get).toHaveBeenCalled();
  expect((axiosWithAuth.get as jest.Mock).mock.calls[0][0]).toEqual(
    `api/kube/apis/packages.operators.coreos.com/v1/namespaces/${ns}/packagemanifests`,
  );
});

it("get operator", async () => {
  const operator = { metadata: { name: "foo" } } as IPackageManifest;
  const ns = "default";
  const opName = "foo";
  axiosWithAuth.get = jest.fn(() => {
    return { data: operator };
  });
  expect(await Operators.getOperator(ns, opName)).toEqual(operator);
  expect(axiosWithAuth.get).toHaveBeenCalled();
  expect((axiosWithAuth.get as jest.Mock).mock.calls[0][0]).toEqual(
    `api/kube/apis/packages.operators.coreos.com/v1/namespaces/${ns}/packagemanifests/${opName}`,
  );
});

it("get csvs", async () => {
  const csv = { metadata: { name: "foo" } } as IClusterServiceVersion;
  const ns = "default";
  axiosWithAuth.get = jest.fn(() => {
    return { data: { items: [csv] } };
  });
  expect(await Operators.getCSVs(ns)).toEqual([csv]);
  expect(axiosWithAuth.get).toHaveBeenCalled();
  expect((axiosWithAuth.get as jest.Mock).mock.calls[0][0]).toEqual(
    `api/kube/apis/operators.coreos.com/v1alpha1/namespaces/${ns}/clusterserviceversions`,
  );
});

it("get global csvs", async () => {
  const csv = { metadata: { name: "foo" } } as IClusterServiceVersion;
  const ns = "_all";
  axiosWithAuth.get = jest.fn(() => {
    return { data: { items: [csv] } };
  });
  expect(await Operators.getCSVs(ns)).toEqual([csv]);
  expect(axiosWithAuth.get).toHaveBeenCalled();
  expect((axiosWithAuth.get as jest.Mock).mock.calls[0][0]).toEqual(
    "api/kube/apis/operators.coreos.com/v1alpha1/namespaces/operators/clusterserviceversions",
  );
});

it("get csv", async () => {
  const csv = { metadata: { name: "foo" } } as IClusterServiceVersion;
  const ns = "default";
  axiosWithAuth.get = jest.fn(() => {
    return { data: csv };
  });
  expect(await Operators.getCSV(ns, "foo")).toEqual(csv);
  expect(axiosWithAuth.get).toHaveBeenCalled();
  expect((axiosWithAuth.get as jest.Mock).mock.calls[0][0]).toEqual(
    `api/kube/apis/operators.coreos.com/v1alpha1/namespaces/${ns}/clusterserviceversions/foo`,
  );
});

it("creates a resource", async () => {
  const resource = { metadata: { name: "foo" } } as IResource;
  const ns = "default";
  axiosWithAuth.post = jest.fn(() => {
    return { data: resource };
  });
  expect(await Operators.createResource(ns, "v1", "pods", resource)).toEqual(resource);
  expect(axiosWithAuth.post).toHaveBeenCalled();
  expect((axiosWithAuth.post as jest.Mock).mock.calls[0]).toEqual([
    `api/kube/apis/v1/namespaces/${ns}/pods`,
    resource,
  ]);
});

it("list resources", async () => {
  const resource = { metadata: { name: "foo" } } as IResource;
  const ns = "default";
  axiosWithAuth.get = jest.fn(() => {
    return { data: { items: [resource] } };
  });
  expect(await Operators.listResources(ns, "v1", "pods")).toEqual({ items: [resource] });
  expect(axiosWithAuth.get).toHaveBeenCalled();
  expect((axiosWithAuth.get as jest.Mock).mock.calls[0]).toEqual([
    `api/kube/apis/v1/namespaces/${ns}/pods`,
  ]);
});

it("get a resource", async () => {
  const resource = { metadata: { name: "foo" } } as IResource;
  const ns = "default";
  axiosWithAuth.get = jest.fn(() => {
    return { data: resource };
  });
  expect(await Operators.getResource(ns, "v1", "pods", "foo")).toEqual(resource);
  expect(axiosWithAuth.get).toHaveBeenCalled();
  expect((axiosWithAuth.get as jest.Mock).mock.calls[0]).toEqual([
    `api/kube/apis/v1/namespaces/${ns}/pods/foo`,
  ]);
});

it("deletes a resource", async () => {
  const resource = { metadata: { name: "foo" } } as IResource;
  const ns = "default";
  axiosWithAuth.delete = jest.fn(() => {
    return { data: resource };
  });
  expect(await Operators.deleteResource(ns, "v1", "pods", "foo")).toEqual(resource);
  expect(axiosWithAuth.delete).toHaveBeenCalled();
  expect((axiosWithAuth.delete as jest.Mock).mock.calls[0]).toEqual([
    `api/kube/apis/v1/namespaces/${ns}/pods/foo`,
  ]);
});

it("updates a resource", async () => {
  const resource = { metadata: { name: "foo" } } as IResource;
  const ns = "default";
  axiosWithAuth.put = jest.fn(() => {
    return { data: resource };
  });
  expect(
    await Operators.updateResource(ns, "v1", "pods", resource.metadata.name, resource),
  ).toEqual(resource);
  expect(axiosWithAuth.post).toHaveBeenCalled();
  expect((axiosWithAuth.post as jest.Mock).mock.calls[0]).toEqual([
    `api/kube/apis/v1/namespaces/${ns}/pods`,
    resource,
  ]);
});
