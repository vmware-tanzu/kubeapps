import * as moxios from "moxios";
import { axiosWithAuth } from "./AxiosInstance";
import { Kube } from "./Kube";

const clusterName = "cluster-name";

describe("App", () => {
  beforeEach(() => {
    // Import as "any" to avoid typescript syntax error
    moxios.install(axiosWithAuth as any);
  });
  afterEach(() => {
    moxios.uninstall(axiosWithAuth as any);
  });
  describe("getResourceURL", () => {
    [
      {
        description: "returns the version and resource",
        args: [clusterName, "v1", "pods"],
        result: `api/clusters/${clusterName}/api/v1/pods`,
      },
      {
        description: "returns the version and resource",
        args: [clusterName, "v1", "pods"],
        result: `api/clusters/${clusterName}/api/v1/pods`,
      },
      {
        description: "returns the version and resource",
        args: [clusterName, "v1", "pods"],
        result: `api/clusters/${clusterName}/api/v1/pods`,
      },
      {
        description: "defaults version to api/v1 if undefined",
        args: [clusterName, "", "pods"],
        result: `api/clusters/${clusterName}/api/v1/pods`,
      },
      {
        description: "returns the version, resource in a namespace",
        args: [clusterName, "v1", "pods", "default"],
        result: `api/clusters/${clusterName}/api/v1/namespaces/default/pods`,
      },
      {
        description: "returns the version, resource in a namespace with a name",
        args: [clusterName, "v1", "pods", "default", "foo"],
        result: `api/clusters/${clusterName}/api/v1/namespaces/default/pods/foo`,
      },
      {
        description: "returns the version, resource in a namespace with a name and a query",
        args: [clusterName, "v1", "pods", "default", "foo", "label=bar"],
        result: `api/clusters/${clusterName}/api/v1/namespaces/default/pods/foo?label=bar`,
      },
    ].forEach(t => {
      it(t.description, () => {
        expect(
          Kube.getResourceURL(t.args[0], t.args[1], t.args[2], t.args[3], t.args[4], t.args[5]),
        ).toBe(t.result);
      });
    });
  });

  describe("watchResourceURL", () => {
    [
      {
        description: "returns the version and resource",
        args: [clusterName, "v1", "pods"],
        result: `ws://localhost/api/clusters/${clusterName}/api/v1/pods?watch=true`,
      },
      {
        description: "returns the version, resource in a namespace",
        args: [clusterName, "v1", "pods", "default"],
        result: `ws://localhost/api/clusters/${clusterName}/api/v1/namespaces/default/pods?watch=true`,
      },
      {
        description: "returns the version, resource in a namespace with a name",
        args: [clusterName, "v1", "pods", "default", "foo"],
        result: `ws://localhost/api/clusters/${clusterName}/api/v1/namespaces/default/pods?watch=true&fieldSelector=metadata.name%3Dfoo`,
      },
      {
        description: "returns the version, resource in a namespace with a name and a query",
        args: [clusterName, "v1", "pods", "default", "foo", "label=bar"],
        result: `ws://localhost/api/clusters/${clusterName}/api/v1/namespaces/default/pods?watch=true&fieldSelector=metadata.name%3Dfoo&label=bar`,
      },
    ].forEach(t => {
      it(t.description, () => {
        expect(
          Kube.watchResourceURL(t.args[0], t.args[1], t.args[2], t.args[3], t.args[4], t.args[5]),
        ).toBe(t.result);
      });
    });
  });

  describe("getResource", () => {
    const resource = { name: "foo" };
    beforeEach(() => {
      moxios.stubRequest(/.*/, {
        response: { data: resource },
        status: 200,
      });
    });
    it("should request a resource", async () => {
      expect(
        await Kube.getResource(clusterName, "v1", "pods", "default", "foo", "label=bar"),
      ).toEqual({
        data: resource,
      });
      expect(moxios.requests.mostRecent().url).toBe(
        `api/clusters/${clusterName}/api/v1/namespaces/default/pods/foo?label=bar`,
      );
    });
  });

  describe("watchResource", () => {
    it("should open a socket", async () => {
      const socket = Kube.watchResource(clusterName, "v1", "pods", "default", "foo");
      expect(socket.url).toBe(
        `ws://localhost/api/clusters/${clusterName}/api/v1/namespaces/default/pods?watch=true&fieldSelector=metadata.name%3Dfoo`,
      );
      // it's a mock socket, so doesn't actually need to be closed
      socket.close();
    });
  });
});
