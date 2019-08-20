import * as moxios from "moxios";
import { App, TILLER_PROXY_ROOT_URL } from "./App";
import { AppRepository } from "./AppRepository";
import { axiosWithAuth } from "./AxiosInstance";
import { IAppOverview, IChartVersion } from "./types";

describe("App", () => {
  beforeEach(() => {
    // Import as "any" to avoid typescript syntax error
    moxios.install(axiosWithAuth as any);
  });
  afterEach(() => {
    moxios.uninstall(axiosWithAuth as any);
    jest.resetAllMocks();
  });
  describe("getResourceURL", () => {
    [
      {
        description: "returns the root API URL if no params are given",
        result: `${TILLER_PROXY_ROOT_URL}/releases`,
      },
      {
        description: "returns namespaced URLs",
        namespace: "default",
        result: `${TILLER_PROXY_ROOT_URL}/namespaces/default/releases`,
      },
      {
        description: "returns a single release URL",
        namespace: "default",
        resourceName: "foo",
        result: `${TILLER_PROXY_ROOT_URL}/namespaces/default/releases/foo`,
      },
      {
        description: "returns a URL with a query",
        namespace: "default",
        query: "statuses=foo",
        result: `${TILLER_PROXY_ROOT_URL}/namespaces/default/releases?statuses=foo`,
      },
    ].forEach(t => {
      it(t.description, () => {
        expect(App.getResourceURL(t.namespace, t.resourceName, t.query)).toBe(t.result);
      });
    });
  });

  describe("listApps", () => {
    const apps = [{ releaseName: "foo" } as IAppOverview];
    beforeEach(() => {
      moxios.stubRequest(/.*/, {
        response: { data: apps },
        status: 200,
      });
    });
    [
      {
        description: "should request all the releases if no namespace is given",
        expectedURL: `${TILLER_PROXY_ROOT_URL}/releases`,
      },
      {
        description: "should request the releases of a namespace",
        expectedURL: `${TILLER_PROXY_ROOT_URL}/namespaces/default/releases`,
        namespace: "default",
      },
      {
        all: true,
        description: "should request the releases of a namespace with any status",
        expectedURL: `${TILLER_PROXY_ROOT_URL}/namespaces/default/releases?statuses=all`,
        namespace: "default",
      },
    ].forEach(t => {
      it(t.description, async () => {
        expect(await App.listApps(t.namespace, t.all)).toEqual(apps);
        expect(moxios.requests.mostRecent().url).toBe(t.expectedURL);
      });
    });
  });
  describe("delete", () => {
    [
      {
        description: "should delete an app in a namespace",
        expectedURL: `${TILLER_PROXY_ROOT_URL}/namespaces/default/releases/foo`,
        purge: false,
      },
      {
        description: "should delete and purge an app in a namespace",
        expectedURL: `${TILLER_PROXY_ROOT_URL}/namespaces/default/releases/foo?purge=true`,
        purge: true,
      },
    ].forEach(t => {
      it(t.description, async () => {
        moxios.stubRequest(/.*/, { response: "ok", status: 200 });
        expect(await App.delete("foo", "default", t.purge)).toBe("ok");
        expect(moxios.requests.mostRecent().url).toBe(t.expectedURL);
      });
    });
    it("throws an error if returns an error 404", async () => {
      moxios.stubRequest(/.*/, { status: 404 });
      let errored = false;
      try {
        await App.delete("foo", "default", false);
      } catch (e) {
        errored = true;
        expect(e.message).toBe("Request failed with status code 404");
      } finally {
        expect(errored).toBe(true);
      }
    });
  });

  describe("rollback", () => {
    it("should rollback an application", async () => {
      AppRepository.get = jest.fn().mockReturnValue({ spec: { auth: {} } });
      axiosWithAuth.put = jest.fn().mockReturnValue({ data: "ok" });
      const c = {
        attributes: { version: "1.0.0" },
        relationships: { chart: { data: { name: "foo", repo: { name: "bar", url: "test.com" } } } },
      } as IChartVersion;
      expect(await App.rollback("foo", "default", 1, "kubeapps", c, "foo: bar")).toBe("ok");
      expect(axiosWithAuth.put).toBeCalledWith(
        "api/tiller-deploy/v1/namespaces/default/releases/foo",
        {
          auth: {},
          chartName: "foo",
          releaseName: "foo",
          repoUrl: "test.com",
          values: "foo: bar",
          version: "1.0.0",
        },
        { params: { action: "rollback", revision: 1 } },
      );
    });
  });
});
