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
    "api/kube/apis/apiextensions.k8s.io/v1/customresourcedefinitions/clusterserviceversions.operators.coreos.com",
  );
});

it("OLM is not installed if the request fails", async () => {
  axiosWithAuth.get = jest.fn(() => {
    throw new Error("nope");
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
