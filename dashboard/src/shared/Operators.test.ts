import { axiosWithAuth } from "./AxiosInstance";
import { Operators } from "./Operators";
import { IPackageManifest } from "./types";

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
