import * as moxios from "moxios";
import { axios } from "./Auth";
import { Kube, KUBE_ROOT_URL } from "./Kube";

describe("App", () => {
  beforeEach(() => {
    moxios.install(axios);
  });
  afterEach(() => {
    moxios.uninstall(axios);
  });
  describe("getResourceURL", () => {
    [
      {
        description: "returns the version and resource",
        args: ["v1", "pods"],
        result: `${KUBE_ROOT_URL}/api/v1/pods`,
      },
      {
        description: "returns the version, resource in a namespace",
        args: ["v1", "pods", "default"],
        result: `${KUBE_ROOT_URL}/api/v1/namespaces/default/pods`,
      },
      {
        description: "returns the version, resource in a namespace with a name",
        args: ["v1", "pods", "default", "foo"],
        result: `${KUBE_ROOT_URL}/api/v1/namespaces/default/pods/foo`,
      },
      {
        description: "returns the version, resource in a namespace with a name and a query",
        args: ["v1", "pods", "default", "foo", "label=bar"],
        result: `${KUBE_ROOT_URL}/api/v1/namespaces/default/pods/foo?label=bar`,
      },
    ].forEach(t => {
      it(t.description, () => {
        expect(Kube.getResourceURL(t.args[0], t.args[1], t.args[2], t.args[3], t.args[4])).toBe(
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
        `${KUBE_ROOT_URL}/api/v1/namespaces/default/pods/foo?label=bar`,
      );
    });
  });
});
