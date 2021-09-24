import {
  AvailablePackageReference,
  Context,
  CreateInstalledPackageResponse,
  DeleteInstalledPackageResponse,
  InstalledPackageReference,
  UpdateInstalledPackageResponse,
  VersionReference,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import * as moxios from "moxios";
import { App } from "./App";
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
        description: "should call to createInstalledPackage",
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

  describe("deleteInstalledPackage", () => {
    [
      {
        description: "should call to deleteInstalledPackage",
        args: {
          installedPackageReference: {
            context: { cluster: "default-c", namespace: "default-ns" },
            identifier: "foo",
            plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
          } as InstalledPackageReference,
        },
      },
    ].forEach(t => {
      it(t.description, async () => {
        const mockDeleteInstalledPackage = jest
          .fn()
          .mockImplementation(() => Promise.resolve({} as DeleteInstalledPackageResponse));
        jest.spyOn(App, "deleteInstalledPackage").mockImplementation(mockDeleteInstalledPackage);
        const res = await App.deleteInstalledPackage(t.args.installedPackageReference);
        expect(res).toStrictEqual({} as DeleteInstalledPackageResponse);
        expect(mockDeleteInstalledPackage).toHaveBeenCalledWith(...Object.values(t.args));
      });
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
