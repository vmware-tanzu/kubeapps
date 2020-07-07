import { axiosWithAuth } from "./AxiosInstance";
import { findOwnedKind, getIcon, Operators } from "./Operators";
import { IClusterServiceVersion, IPackageManifest, IResource } from "./types";

it("check if the OLM has been installed", async () => {
  axiosWithAuth.get = jest.fn().mockReturnValue({ status: 200 });
  expect(await Operators.isOLMInstalled("ns")).toBe(true);
  expect(axiosWithAuth.get).toHaveBeenCalled();
  expect((axiosWithAuth.get as jest.Mock).mock.calls[0][0]).toEqual(
    "api/clusters/default/apis/packages.operators.coreos.com/v1/namespaces/ns/packagemanifests",
  );
});

it("OLM is not installed if the request fails", async () => {
  axiosWithAuth.get = jest.fn().mockReturnValue({ status: 404 });
  expect(await Operators.isOLMInstalled("ns")).toBe(false);
});

it("OLM is not installed if the request returns != 200", async () => {
  axiosWithAuth.get = jest.fn().mockReturnValue({ status: 404 });
  expect(await Operators.isOLMInstalled("ns")).toBe(false);
});

it("get operators", async () => {
  const operator = { metadata: { name: "foo" } } as IPackageManifest;
  const ns = "default";
  axiosWithAuth.get = jest.fn().mockReturnValue({ data: { items: [operator] } });
  expect(await Operators.getOperators(ns)).toEqual([operator]);
  expect(axiosWithAuth.get).toHaveBeenCalled();
  expect((axiosWithAuth.get as jest.Mock).mock.calls[0][0]).toEqual(
    `api/clusters/default/apis/packages.operators.coreos.com/v1/namespaces/${ns}/packagemanifests`,
  );
});

it("get operator", async () => {
  const operator = { metadata: { name: "foo" } } as IPackageManifest;
  const ns = "default";
  const opName = "foo";
  axiosWithAuth.get = jest.fn().mockReturnValue({ data: operator });
  expect(await Operators.getOperator(ns, opName)).toEqual(operator);
  expect(axiosWithAuth.get).toHaveBeenCalled();
  expect((axiosWithAuth.get as jest.Mock).mock.calls[0][0]).toEqual(
    `api/clusters/default/apis/packages.operators.coreos.com/v1/namespaces/${ns}/packagemanifests/${opName}`,
  );
});

it("get csvs", async () => {
  const csv = { metadata: { name: "foo" } } as IClusterServiceVersion;
  const ns = "default";
  axiosWithAuth.get = jest.fn().mockReturnValue({ data: { items: [csv] } });
  expect(await Operators.getCSVs(ns)).toEqual([csv]);
  expect(axiosWithAuth.get).toHaveBeenCalled();
  expect((axiosWithAuth.get as jest.Mock).mock.calls[0][0]).toEqual(
    `api/clusters/default/apis/operators.coreos.com/v1alpha1/namespaces/${ns}/clusterserviceversions`,
  );
});

it("get global csvs", async () => {
  const csv = { metadata: { name: "foo" } } as IClusterServiceVersion;
  const ns = "_all";
  axiosWithAuth.get = jest.fn().mockReturnValue({ data: { items: [csv] } });
  expect(await Operators.getCSVs(ns)).toEqual([csv]);
  expect(axiosWithAuth.get).toHaveBeenCalled();
  expect((axiosWithAuth.get as jest.Mock).mock.calls[0][0]).toEqual(
    "api/clusters/default/apis/operators.coreos.com/v1alpha1/namespaces/operators/clusterserviceversions",
  );
});

it("get csv", async () => {
  const csv = { metadata: { name: "foo" } } as IClusterServiceVersion;
  const ns = "default";
  axiosWithAuth.get = jest.fn().mockReturnValue({ data: csv });
  expect(await Operators.getCSV(ns, "foo")).toEqual(csv);
  expect(axiosWithAuth.get).toHaveBeenCalled();
  expect((axiosWithAuth.get as jest.Mock).mock.calls[0][0]).toEqual(
    `api/clusters/default/apis/operators.coreos.com/v1alpha1/namespaces/${ns}/clusterserviceversions/foo`,
  );
});

it("creates a resource", async () => {
  const resource = { metadata: { name: "foo" } } as IResource;
  const ns = "default";
  axiosWithAuth.post = jest.fn().mockReturnValue({ data: resource });
  expect(await Operators.createResource(ns, "v1", "pods", resource)).toEqual(resource);
  expect(axiosWithAuth.post).toHaveBeenCalled();
  expect((axiosWithAuth.post as jest.Mock).mock.calls[0]).toEqual([
    `api/clusters/default/apis/v1/namespaces/${ns}/pods`,
    resource,
  ]);
});

it("list resources", async () => {
  const resource = { metadata: { name: "foo" } } as IResource;
  const ns = "default";
  axiosWithAuth.get = jest.fn().mockReturnValue({ data: { items: [resource] } });
  expect(await Operators.listResources(ns, "v1", "pods")).toEqual({ items: [resource] });
  expect(axiosWithAuth.get).toHaveBeenCalled();
  expect((axiosWithAuth.get as jest.Mock).mock.calls[0]).toEqual([
    `api/clusters/default/apis/v1/namespaces/${ns}/pods`,
  ]);
});

it("get a resource", async () => {
  const resource = { metadata: { name: "foo" } } as IResource;
  const ns = "default";
  axiosWithAuth.get = jest.fn().mockReturnValue({ data: resource });
  expect(await Operators.getResource(ns, "v1", "pods", "foo")).toEqual(resource);
  expect(axiosWithAuth.get).toHaveBeenCalled();
  expect((axiosWithAuth.get as jest.Mock).mock.calls[0]).toEqual([
    `api/clusters/default/apis/v1/namespaces/${ns}/pods/foo`,
  ]);
});

it("deletes a resource", async () => {
  const resource = { metadata: { name: "foo" } } as IResource;
  const ns = "default";
  axiosWithAuth.delete = jest.fn().mockReturnValue({ data: resource });
  expect(await Operators.deleteResource(ns, "v1", "pods", "foo")).toEqual(resource);
  expect(axiosWithAuth.delete).toHaveBeenCalled();
  expect((axiosWithAuth.delete as jest.Mock).mock.calls[0]).toEqual([
    `api/clusters/default/apis/v1/namespaces/${ns}/pods/foo`,
  ]);
});

it("updates a resource", async () => {
  const resource = { metadata: { name: "foo" } } as IResource;
  const ns = "default";
  axiosWithAuth.put = jest.fn().mockReturnValue({ data: resource });
  expect(
    await Operators.updateResource(ns, "v1", "pods", resource.metadata.name, resource),
  ).toEqual(resource);
  expect(axiosWithAuth.post).toHaveBeenCalled();
  expect((axiosWithAuth.post as jest.Mock).mock.calls[0]).toEqual([
    `api/clusters/default/apis/v1/namespaces/${ns}/pods`,
    resource,
  ]);
});

const namespace = "default";
const subscription = {
  apiVersion: "operators.coreos.com/v1alpha1",
  kind: "Subscription",
  metadata: {
    name: "foo",
    namespace,
  },
  spec: {
    channel: "alpha",
    installPlanApproval: "Manual",
    name: "foo",
    source: "operatorhubio-catalog",
    sourceNamespace: "olm",
    startingCSV: "foo.1.0.0",
  },
};
const operatorgroup = {
  apiVersion: "operators.coreos.com/v1",
  kind: "OperatorGroup",
  metadata: {
    generateName: "default-",
    namespace,
  },
  spec: {
    targetNamespaces: [namespace],
  },
};

it("creates an operatorgroup and a subscription", async () => {
  const operatorGroups = { items: [] };
  axiosWithAuth.get = jest.fn().mockReturnValue({ data: operatorGroups });
  const resource = { metadata: { name: "foo" } } as IResource;
  axiosWithAuth.post = jest.fn().mockReturnValue({ data: resource });
  expect(await Operators.createOperator(namespace, "foo", "alpha", "Manual", "foo.1.0.0")).toEqual(
    resource,
  );
  expect(axiosWithAuth.get).toHaveBeenCalled();
  expect((axiosWithAuth.get as jest.Mock).mock.calls[0]).toEqual([
    `api/clusters/default/apis/operators.coreos.com/v1/namespaces/${namespace}/operatorgroups`,
  ]);
  expect(axiosWithAuth.post).toHaveBeenCalledTimes(2);
  expect((axiosWithAuth.post as jest.Mock).mock.calls[0]).toEqual([
    `api/clusters/default/apis/operators.coreos.com/v1/namespaces/${namespace}/operatorgroups`,
    operatorgroup,
  ]);
  expect((axiosWithAuth.post as jest.Mock).mock.calls[1]).toEqual([
    `api/clusters/default/apis/operators.coreos.com/v1alpha1/namespaces/${namespace}/subscriptions/foo`,
    subscription,
  ]);
});

it("creates only a subscription if the operator group already exists", async () => {
  const operatorGroups = { items: [{ metadata: { name: "foo" } }] };
  axiosWithAuth.get = jest.fn().mockReturnValue({ data: operatorGroups });
  const resource = { metadata: { name: "foo" } } as IResource;
  axiosWithAuth.post = jest.fn().mockReturnValue({ data: resource });
  expect(await Operators.createOperator(namespace, "foo", "alpha", "Manual", "foo.1.0.0")).toEqual(
    resource,
  );
  expect(axiosWithAuth.get).toHaveBeenCalled();
  expect((axiosWithAuth.get as jest.Mock).mock.calls[0]).toEqual([
    `api/clusters/default/apis/operators.coreos.com/v1/namespaces/${namespace}/operatorgroups`,
  ]);
  expect(axiosWithAuth.post).toHaveBeenCalledTimes(1);
  expect((axiosWithAuth.post as jest.Mock).mock.calls[0]).toEqual([
    `api/clusters/default/apis/operators.coreos.com/v1alpha1/namespaces/${namespace}/subscriptions/foo`,
    subscription,
  ]);
});

it("creates only a subscription if the namespace is operators", async () => {
  const operatorGroups = { items: [] };
  axiosWithAuth.get = jest.fn().mockReturnValue({ data: operatorGroups });
  const resource = { metadata: { name: "foo" } } as IResource;
  axiosWithAuth.post = jest.fn().mockReturnValue({ data: resource });
  expect(
    await Operators.createOperator("operators", "foo", "alpha", "Manual", "foo.1.0.0"),
  ).toEqual(resource);
  expect(axiosWithAuth.get).not.toHaveBeenCalled();
  expect(axiosWithAuth.post).toHaveBeenCalledTimes(1);
});

it("finds a default channel", () => {
  const operator = {
    status: {
      defaultChannel: "foo",
      channels: [{ name: "foo" }, { name: "bar" }],
    },
  } as any;
  expect(Operators.getDefaultChannel(operator)).toEqual({ name: "foo" });
});

describe("#global", () => {
  [
    {
      description: "returns true if the channel support all namespaces",
      channel: { currentCSVDesc: { installModes: [{ type: "AllNamespaces", supported: true }] } },
      result: true,
    },
    {
      description: "returns false if the channel support only namespaces",
      channel: { currentCSVDesc: { installModes: [{ type: "AllNamespaces", supported: false }] } },
      result: false,
    },
    {
      description: "returns false if the channel is undefined",
      channel: undefined,
      result: false,
    },
  ].forEach(test => {
    it(test.description, () => {
      expect(Operators.global(test.channel as any)).toBe(test.result);
    });
  });
});

describe("#getIcon", () => {
  it("extracts an icon from a csv", () => {
    const csv = {
      spec: { icon: [{ mediatype: "foo", base64data: "bar" }] },
    } as IClusterServiceVersion;
    expect(getIcon(csv)).toEqual("data:foo;base64,bar");
  });

  it("returns a placeholder if no info is found", () => {
    const csv = {
      spec: {},
    } as IClusterServiceVersion;
    expect(getIcon(csv)).toEqual("placeholder.png");
  });
});

describe("#findOwnedKind", () => {
  it("finds an owned kind", () => {
    const csv = {
      spec: {
        customresourcedefinitions: {
          owned: [
            {
              kind: "foo",
            },
          ],
        },
      },
    } as IClusterServiceVersion;
    expect(findOwnedKind(csv, "foo")).toEqual({ kind: "foo" });
  });
});
