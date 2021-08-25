import { AvailablePackageDetail } from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import * as moxios from "moxios";
import { App, KUBEOPS_ROOT_URL } from "./App";
import { axiosWithAuth } from "./AxiosInstance";

const availablePackageDetail: AvailablePackageDetail = {
  name: "testrepo/test",
  categories: [],
  defaultValues: "",
  displayName: "test",
  homeUrl: "",
  iconUrl: "",
  longDescription: "bla bla",
  maintainers: [],
  readme: "",
  repoUrl: "http://example.com/charts",
  shortDescription: "bla",
  sourceUrls: [""],
  valuesSchema: "",
  version: { appVersion: "10.0.0", pkgVersion: "1.0.0" },
  availablePackageRef: {
    identifier: "default",
    context: { cluster: "default", namespace: "kubeapps" },
  },
};

describe("App", () => {
  beforeEach(() => {
    // Import as "any" to avoid typescript syntax error
    moxios.install(axiosWithAuth as any);
  });
  afterEach(() => {
    moxios.uninstall(axiosWithAuth as any);
    jest.resetAllMocks();
  });

  // TODO(agamez): Test temporarily commented out
  // describe("listApps", () => {
  //   const apps = [{ name: "foo" } as InstalledPackageSummary];
  //   beforeEach(() => {
  //     moxios.stubRequest(/.*/, {
  //       response: { data: apps },
  //       status: 200,
  //     });
  //   });
  //   [
  //     {
  //       description: "should request all the releases if no namespace is given",
  //       expectedURL: `${KUBEOPS_ROOT_URL}/clusters/defaultc/releases?statuses=all`,
  //     },
  //     {
  //       description: "should request the releases of a namespace with any status",
  //       expectedURL: `${KUBEOPS_ROOT_URL}/clusters/defaultc/namespaces/default/releases?statuses=all`,
  //       namespace: "default",
  //     },
  //   ].forEach(t => {
  //     it(t.description, async () => {
  //       expect(await App.GetInstalledPackageSummaries("defaultc", t.namespace)).toEqual(apps);
  //       expect(moxios.requests.mostRecent().url).toBe(t.expectedURL);
  //     });
  //   });
  // });

  describe("create", () => {
    const expectedURL = `${KUBEOPS_ROOT_URL}/clusters/defaultc/namespaces/defaultns/releases`;

    it("creates an app in a namespace", async () => {
      moxios.stubRequest(/.*/, { response: "ok", status: 200 });
      expect(await App.create("defaultc", "defaultns", "absent-ant", availablePackageDetail)).toBe(
        "ok",
      );
      const request = moxios.requests.mostRecent();
      expect(request.url).toBe(expectedURL);
      expect(request.config.data).toEqual(
        JSON.stringify({
          appRepositoryResourceName:
            availablePackageDetail.availablePackageRef?.identifier.split("/")[0],
          appRepositoryResourceNamespace:
            availablePackageDetail.availablePackageRef?.context?.namespace,
          chartName: decodeURIComponent(availablePackageDetail.name),
          releaseName: "absent-ant",
          version: availablePackageDetail.version?.pkgVersion,
        }),
      );
    });
  });

  describe("upgrade", () => {
    const expectedURL = `${KUBEOPS_ROOT_URL}/clusters/default-c/namespaces/default-ns/releases/absent-ant`;
    it("upgrades an app in a namespace", async () => {
      moxios.stubRequest(/.*/, { response: "ok", status: 200 });
      expect(
        await App.upgrade(
          "default-c",
          "default-ns",
          "absent-ant",
          "kubeapps",
          availablePackageDetail,
        ),
      ).toBe("ok");
      const request = moxios.requests.mostRecent();
      expect(request.url).toBe(expectedURL);
      expect(request.config.data).toEqual(
        JSON.stringify({
          appRepositoryResourceName:
            availablePackageDetail.availablePackageRef?.identifier.split("/")[0],
          appRepositoryResourceNamespace: "kubeapps",
          chartName: decodeURIComponent(availablePackageDetail.name),
          releaseName: "absent-ant",
          version: availablePackageDetail.version?.pkgVersion,
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
