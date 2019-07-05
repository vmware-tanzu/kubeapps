import * as moxios from "moxios";
import { axiosWithAuth } from "./AxiosInstance";
import { APIBase, Kube, WebSocketAPIBase } from "./Kube";

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
        args: ["v1", "pods"],
        result: `${APIBase}/api/v1/pods`,
      },
      {
        description: "returns the version, resource in a namespace",
        args: ["v1", "pods", "default"],
        result: `${APIBase}/api/v1/namespaces/default/pods`,
      },
      {
        description: "returns the version, resource in a namespace with a name",
        args: ["v1", "pods", "default", "foo"],
        result: `${APIBase}/api/v1/namespaces/default/pods/foo`,
      },
      {
        description: "returns the version, resource in a namespace with a name and a query",
        args: ["v1", "pods", "default", "foo", "label=bar"],
        result: `${APIBase}/api/v1/namespaces/default/pods/foo?label=bar`,
      },
    ].forEach(t => {
      it(t.description, () => {
        expect(Kube.getResourceURL(t.args[0], t.args[1], t.args[2], t.args[3], t.args[4])).toBe(
          t.result,
        );
      });
    });
  });

  describe("watchResourceURL", () => {
    [
      {
        description: "returns the version and resource",
        args: ["v1", "pods"],
        result: `${WebSocketAPIBase}${APIBase}/api/v1/pods?watch=true`,
      },
      {
        description: "returns the version, resource in a namespace",
        args: ["v1", "pods", "default"],
        result: `${WebSocketAPIBase}${APIBase}/api/v1/namespaces/default/pods?watch=true`,
      },
      {
        description: "returns the version, resource in a namespace with a name",
        args: ["v1", "pods", "default", "foo"],
        result: `${WebSocketAPIBase}${APIBase}/api/v1/namespaces/default/pods?watch=true&fieldSelector=metadata.name%3Dfoo`,
      },
      {
        description: "returns the version, resource in a namespace with a name and a query",
        args: ["v1", "pods", "default", "foo", "label=bar"],
        result: `${WebSocketAPIBase}${APIBase}/api/v1/namespaces/default/pods?watch=true&fieldSelector=metadata.name%3Dfoo&label=bar`,
      },
    ].forEach(t => {
      it(t.description, () => {
        expect(Kube.watchResourceURL(t.args[0], t.args[1], t.args[2], t.args[3], t.args[4])).toBe(
          t.result,
        );
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
      expect(await Kube.getResource("v1", "pods", "default", "foo", "label=bar")).toEqual({
        data: resource,
      });
      expect(moxios.requests.mostRecent().url).toBe(
        `${APIBase}/api/v1/namespaces/default/pods/foo?label=bar`,
      );
    });
  });

  describe("watchResource", () => {
    it("should open a socket", async () => {
      const socket = Kube.watchResource("v1", "pods", "default", "foo");
      expect(socket.url).toBe(
        `${WebSocketAPIBase}${APIBase}/api/v1/namespaces/default/pods?watch=true&fieldSelector=metadata.name%3Dfoo`,
      );
      // it's a mock socket, so doesn't actually need to be closed
      socket.close();
    });
  });

  describe("resourcePlural", () => {
    const tests = [
      { kind: "Service", expected: "services" },
      { kind: "Ingress", expected: "ingresses" },
      { kind: "Deployment", expected: "deployments" },
    ];
    tests.forEach(t => {
      it(`returns the correct plural for ${t.kind}`, () => {
        expect(Kube.resourcePlural(t.kind)).toBe(t.expected);
      });
    });

    it("throws an error if the resource kind isn't registered", () => {
      expect(() => Kube.resourcePlural("ThisKindWillNeverExist")).toThrow(
        "Don't know plural for ThisKindWillNeverExist, register it in Kube",
      );
    });
  });
});
