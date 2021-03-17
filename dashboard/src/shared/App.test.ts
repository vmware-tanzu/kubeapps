import * as moxios from "moxios";
import { App, KUBEOPS_ROOT_URL } from "./App";
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
        expectedURL: `${KUBEOPS_ROOT_URL}/clusters/defaultc/releases?statuses=all`,
      },
      {
        description: "should request the releases of a namespace with any status",
        expectedURL: `${KUBEOPS_ROOT_URL}/clusters/defaultc/namespaces/default/releases?statuses=all`,
        namespace: "default",
      },
    ].forEach(t => {
      it(t.description, async () => {
        expect(await App.listApps("defaultc", t.namespace)).toEqual(apps);
        expect(moxios.requests.mostRecent().url).toBe(t.expectedURL);
      });
    });
  });

  describe("create", () => {
    const expectedURL = `${KUBEOPS_ROOT_URL}/clusters/defaultc/namespaces/defaultns/releases`;
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
      expect(
        await App.create("defaultc", "defaultns", "absent-ant", "kubeapps", testChartVersion),
      ).toBe("ok");
      const request = moxios.requests.mostRecent();
      expect(request.url).toBe(expectedURL);
      expect(request.config.data).toEqual(
        JSON.stringify({
          appRepositoryResourceName: testChartVersion.relationships.chart.data.repo.name,
          appRepositoryResourceNamespace: "kubeapps",
          chartName: testChartVersion.relationships.chart.data.name,
          releaseName: "absent-ant",
          version: testChartVersion.attributes.version,
        }),
      );
    });
  });

  describe("upgrade", () => {
    const expectedURL = `${KUBEOPS_ROOT_URL}/clusters/default-c/namespaces/default-ns/releases/absent-ant`;
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
      expect(
        await App.upgrade("default-c", "default-ns", "absent-ant", "kubeapps", testChartVersion),
      ).toBe("ok");
      const request = moxios.requests.mostRecent();
      expect(request.url).toBe(expectedURL);
      expect(request.config.data).toEqual(
        JSON.stringify({
          appRepositoryResourceName: testChartVersion.relationships.chart.data.repo.name,
          appRepositoryResourceNamespace: "kubeapps",
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
        expectedURL: `${KUBEOPS_ROOT_URL}/clusters/default-c/namespaces/default-ns/releases/foo`,
        purge: false,
      },
      {
        description: "should delete and purge an app in a namespace",
        expectedURL: `${KUBEOPS_ROOT_URL}/clusters/default-c/namespaces/default-ns/releases/foo?purge=true`,
        purge: true,
      },
    ].forEach(t => {
      it(t.description, async () => {
        moxios.stubRequest(/.*/, { response: "ok", status: 200 });
        expect(await App.delete("default-c", "default-ns", "foo", t.purge)).toBe("ok");
        expect(moxios.requests.mostRecent().url).toBe(t.expectedURL);
      });
    });
    it("throws an error if returns an error 404", async () => {
      moxios.stubRequest(/.*/, { status: 404 });
      await expect(App.delete("default-c", "default-ns", "foo", false)).rejects.toThrow(
        "Request failed with status code 404",
      );
    });
  });

  describe("rollback", () => {
    it("should rollback an application", async () => {
      axiosWithAuth.put = jest.fn().mockReturnValue({ data: "ok" });
      expect(await App.rollback("default-c", "default-ns", "foo", 1)).toBe("ok");
      expect(axiosWithAuth.put).toBeCalledWith(
        "api/kubeops/v1/clusters/default-c/namespaces/default-ns/releases/foo",
        {},
        { params: { action: "rollback", revision: 1 } },
      );
    });
  });
});
