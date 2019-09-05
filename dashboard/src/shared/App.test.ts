import * as moxios from "moxios";
import { App, TILLER_PROXY_ROOT_URL } from "./App";
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

  describe("create", () => {
    const expectedURL = `${TILLER_PROXY_ROOT_URL}/namespaces/default/releases`;
    const testChartVersion: IChartVersion = {
      attributes: {
        version: "1.2.3",
      },
      relationships: {
        chart: {
          data: {
            name: "test",
            repo: {
              name: "testrepo",
              url: "http://example.com/charts",
            },
          },
        },
      },
    } as IChartVersion;
    it("creates an app in a namespace", async () => {
      moxios.stubRequest(/.*/, { response: "ok", status: 200 });
      expect(await App.create("absent-ant", "default", "kubeapps", testChartVersion)).toBe("ok");
      const request = moxios.requests.mostRecent();
      expect(request.url).toBe(expectedURL);
      expect(request.config.data).toEqual(
        JSON.stringify({
          appRepositoryResourceName: testChartVersion.relationships.chart.data.repo.name,
          chartName: testChartVersion.relationships.chart.data.name,
          releaseName: "absent-ant",
          version: testChartVersion.attributes.version,
        }),
      );
    });
  });

  describe("upgrade", () => {
    const expectedURL = `${TILLER_PROXY_ROOT_URL}/namespaces/default/releases/absent-ant`;
    const testChartVersion: IChartVersion = {
      attributes: {
        version: "1.2.3",
      },
      relationships: {
        chart: {
          data: {
            name: "test",
            repo: {
              name: "testrepo",
              url: "http://example.com/charts",
            },
          },
        },
      },
    } as IChartVersion;
    it("upgrades an app in a namespace", async () => {
      moxios.stubRequest(/.*/, { response: "ok", status: 200 });
      expect(await App.upgrade("absent-ant", "default", "kubeapps", testChartVersion)).toBe("ok");
      const request = moxios.requests.mostRecent();
      expect(request.url).toBe(expectedURL);
      expect(request.config.data).toEqual(
        JSON.stringify({
          appRepositoryResourceName: testChartVersion.relationships.chart.data.repo.name,
          chartName: testChartVersion.relationships.chart.data.name,
          releaseName: "absent-ant",
          version: testChartVersion.attributes.version,
        }),
      );
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
      axiosWithAuth.put = jest.fn().mockReturnValue({ data: "ok" });
      expect(await App.rollback("foo", "default", 1)).toBe("ok");
      expect(axiosWithAuth.put).toBeCalledWith(
        "api/tiller-deploy/v1/namespaces/default/releases/foo",
        {},
        { params: { action: "rollback", revision: 1 } },
      );
    });
  });
});
