import {
  AvailablePackageReference,
  Context,
  CreateInstalledPackageResponse,
  InstalledPackageReference,
  UpdateInstalledPackageResponse,
  VersionReference,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import * as moxios from "moxios";
import { App, KUBEOPS_ROOT_URL } from "./App";
import { axiosWithAuth } from "./AxiosInstance";

describe("App", () => {
  beforeEach(() => {
    // Import as "any" to avoid typescript syntax error
    moxios.install(axiosWithAuth as any);
  });
  afterEach(() => {
    moxios.uninstall(axiosWithAuth as any);
    jest.restoreAllMocks();
  });

  describe("createInstalledPackage", () => {
    [
      {
        description: "createInstalledPackage basic",
        args: {
          tagetContext: { cluster: "my-cluster", namespace: "my-namespace" } as Context,
          name: "",
          availablePackageRef: {
            identifier: "foo/bar",
            context: { cluster: "my-cluster", namespace: "my-namespace" },
            plugin: { name: "my.plugin", version: "0.0.1" },
          } as AvailablePackageReference,
          pkgVersionReference: { version: "1.2.3" } as VersionReference,
          values: "",
          reconciliationOptions: undefined,
        },
      },
    ].forEach(t => {
      it(t.description, async () => {
        const mockCreateInstalledPackage = jest.fn().mockImplementation(() =>
          Promise.resolve({
            installedPackageRef: {
              identifier: "foo/bar",
              context: { cluster: "my-cluster", namespace: "my-namespace" },
              plugin: { name: "my.plugin", version: "0.0.1" },
            } as InstalledPackageReference,
          } as CreateInstalledPackageResponse),
        );
        jest.spyOn(App, "createInstalledPackage").mockImplementation(mockCreateInstalledPackage);
        const availablePackageSummaries = await App.createInstalledPackage(
          t.args.tagetContext,
          t.args.name,
          t.args.availablePackageRef,
          t.args.pkgVersionReference,
          t.args.values,
          t.args.reconciliationOptions,
        );
        expect(availablePackageSummaries).toStrictEqual({
          installedPackageRef: {
            identifier: "foo/bar",
            context: { cluster: "my-cluster", namespace: "my-namespace" },
            plugin: { name: "my.plugin", version: "0.0.1" },
          } as InstalledPackageReference,
        } as CreateInstalledPackageResponse);
        expect(mockCreateInstalledPackage).toHaveBeenCalledWith(...Object.values(t.args));
      });
    });
  });

  describe("UpdateInstalledPackage", () => {
    [
      {
        description: "should call to updateInstalledPackage",
        args: {
          installedPackageRef: {
            context: { cluster: "default-c", namespace: "default-ns" },
            identifier: "foo",
            plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
          } as InstalledPackageReference,
          pkgVersionReference: { version: "1.2.3" } as VersionReference,
          values: "",
          reconciliationOptions: undefined,
        },
      },
    ].forEach(t => {
      it(t.description, async () => {
        const mockUpdateInstalledPackage = jest.fn().mockImplementation(() =>
          Promise.resolve({
            installedPackageRef: {
              identifier: "foo/bar",
              context: { cluster: "my-cluster", namespace: "my-namespace" },
              plugin: { name: "my.plugin", version: "0.0.1" },
            } as InstalledPackageReference,
          } as UpdateInstalledPackageResponse),
        );
        jest.spyOn(App, "updateInstalledPackage").mockImplementation(mockUpdateInstalledPackage);
        const availablePackageSummaries = await App.updateInstalledPackage(
          t.args.installedPackageRef,
          t.args.pkgVersionReference,
          t.args.values,
          t.args.reconciliationOptions,
        );
        expect(availablePackageSummaries).toStrictEqual({
          installedPackageRef: {
            context: { cluster: "my-cluster", namespace: "my-namespace" },
            identifier: "foo/bar",
            plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
          } as InstalledPackageReference,
        } as UpdateInstalledPackageResponse);
        expect(mockUpdateInstalledPackage).toHaveBeenCalledWith(...Object.values(t.args));
      });
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
        expect(
          await App.delete(
            {
              context: { cluster: "default-c", namespace: "default-ns" },
              identifier: "foo",
              plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
            } as InstalledPackageReference,
            t.purge,
          ),
        ).toBe("ok");
        expect(moxios.requests.mostRecent().url).toBe(t.expectedURL);
      });
    });
    it("throws an error if returns an error 404", async () => {
      moxios.stubRequest(/.*/, { status: 404 });
      await expect(
        App.delete(
          {
            context: { cluster: "default-c", namespace: "default-ns" },
            identifier: "foo",
            plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
          } as InstalledPackageReference,
          false,
        ),
      ).rejects.toThrow("Request failed with status code 404");
    });
  });

  describe("rollback", () => {
    it("should rollback an application", async () => {
      axiosWithAuth.put = jest.fn().mockReturnValue({ data: "ok" });
      expect(
        await App.rollback(
          {
            context: { cluster: "default-c", namespace: "default-ns" },
            identifier: "foo",
            plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
          } as InstalledPackageReference,
          1,
        ),
      ).toBe("ok");
      expect(axiosWithAuth.put).toBeCalledWith(
        "api/kubeops/v1/clusters/default-c/namespaces/default-ns/releases/foo",
        {},
        { params: { action: "rollback", revision: 1 } },
      );
    });
  });
});
