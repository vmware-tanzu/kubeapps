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
import { RollbackInstalledPackageResponse } from "gen/kubeappsapis/plugins/helm/packages/v1alpha1/helm";
import * as moxios from "moxios";
import { App } from "./App";
import { axiosWithAuth } from "./AxiosInstance";
import { PluginNames } from "./utils";

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
        jest.spyOn(App, "CreateInstalledPackage").mockImplementation(mockCreateInstalledPackage);
        const availablePackageSummaries = await App.CreateInstalledPackage(
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

  describe("updateInstalledPackage", () => {
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
        jest.spyOn(App, "UpdateInstalledPackage").mockImplementation(mockUpdateInstalledPackage);
        const availablePackageSummaries = await App.UpdateInstalledPackage(
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
        jest.spyOn(App, "DeleteInstalledPackage").mockImplementation(mockDeleteInstalledPackage);
        const res = await App.DeleteInstalledPackage(t.args.installedPackageReference);
        expect(res).toStrictEqual({} as DeleteInstalledPackageResponse);
        expect(mockDeleteInstalledPackage).toHaveBeenCalledWith(...Object.values(t.args));
      });
    });
  });
});

describe("rollbackInstalledPackage", () => {
  [
    {
      description: "should call to rollbackInstalledPackage",
      args: {
        installedPackageReference: {
          context: { cluster: "default-c", namespace: "default-ns" },
          identifier: "foo",
          plugin: { name: PluginNames.PACKAGES_HELM, version: "0.0.1" } as Plugin,
        } as InstalledPackageReference,
        revision: 1,
      },
    },
  ].forEach(t => {
    it(t.description, async () => {
      const mockRollbackInstalledPackage = jest
        .fn()
        .mockImplementation(() => Promise.resolve({} as RollbackInstalledPackageResponse));
      jest.spyOn(App, "RollbackInstalledPackage").mockImplementation(mockRollbackInstalledPackage);
      const res = await App.RollbackInstalledPackage(
        t.args.installedPackageReference,
        t.args.revision,
      );
      expect(res).toStrictEqual({} as RollbackInstalledPackageResponse);
      expect(mockRollbackInstalledPackage).toHaveBeenCalledWith(...Object.values(t.args));
    });
  });
});
