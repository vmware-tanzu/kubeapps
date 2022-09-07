// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import {
  AvailablePackageReference,
  Context,
  CreateInstalledPackageResponse,
  DeleteInstalledPackageResponse,
  GetInstalledPackageResourceRefsResponse,
  InstalledPackageReference,
  UpdateInstalledPackageResponse,
  VersionReference,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import { RollbackInstalledPackageResponse } from "gen/kubeappsapis/plugins/helm/packages/v1alpha1/helm";
import { InstalledPackage } from "./InstalledPackage";
import { KubeappsGrpcClient } from "./KubeappsGrpcClient";
import { PluginNames } from "./types";

describe("InstalledPackage", () => {
  describe("createInstalledPackage", () => {
    [
      {
        description: "should call to createInstalledPackage",
        args: {
          targetContext: { cluster: "my-cluster", namespace: "my-namespace" } as Context,
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
        setMockCoreClient("CreateInstalledPackage", mockCreateInstalledPackage);

        const availablePackageSummaries = await InstalledPackage.CreateInstalledPackage(
          t.args.targetContext,
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
        expect(mockCreateInstalledPackage).toHaveBeenCalledWith(t.args);
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
        setMockCoreClient("UpdateInstalledPackage", mockUpdateInstalledPackage);

        const availablePackageSummaries = await InstalledPackage.UpdateInstalledPackage(
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
        expect(mockUpdateInstalledPackage).toHaveBeenCalledWith(t.args);
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
        jest
          .spyOn(InstalledPackage, "DeleteInstalledPackage")
          .mockImplementation(mockDeleteInstalledPackage);
        const res = await InstalledPackage.DeleteInstalledPackage(t.args.installedPackageReference);
        expect(res).toStrictEqual({} as DeleteInstalledPackageResponse);
        expect(mockDeleteInstalledPackage).toHaveBeenCalledWith(...Object.values(t.args));
      });
    });
  });

  describe("rollbackInstalledPackage", () => {
    [
      {
        description: "should call to rollbackInstalledPackage",
        args: {
          installedPackageRef: {
            context: { cluster: "default-c", namespace: "default-ns" },
            identifier: "foo",
            plugin: { name: PluginNames.PACKAGES_HELM, version: "0.0.1" } as Plugin,
          } as InstalledPackageReference,
          releaseRevision: 1,
        },
      },
    ].forEach(t => {
      it(t.description, async () => {
        const mockRollbackInstalledPackage = jest
          .fn()
          .mockImplementation(() => Promise.resolve({} as RollbackInstalledPackageResponse));
        setMockHelmClient("RollbackInstalledPackage", mockRollbackInstalledPackage);

        const res = await InstalledPackage.RollbackInstalledPackage(
          t.args.installedPackageRef,
          t.args.releaseRevision,
        );

        expect(res).toStrictEqual({} as RollbackInstalledPackageResponse);
        expect(mockRollbackInstalledPackage).toHaveBeenCalledWith(t.args);
      });
    });
  });

  describe("GetInstalledPackageResourceRefs", () => {
    const installedPackageReference = {
      context: { cluster: "default-c", namespace: "default-ns" },
      identifier: "foo",
      plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
    } as InstalledPackageReference;

    it("returns the resource references", async () => {
      const mockClientGetInstalledPackageResourceRefs = jest
        .fn()
        .mockImplementation(() => Promise.resolve({} as GetInstalledPackageResourceRefsResponse));

      setMockCoreClient(
        "GetInstalledPackageResourceRefs",
        mockClientGetInstalledPackageResourceRefs,
      );

      const res = await InstalledPackage.GetInstalledPackageResourceRefs(installedPackageReference);

      expect(res).toStrictEqual({} as GetInstalledPackageResourceRefsResponse);
      expect(mockClientGetInstalledPackageResourceRefs).toHaveBeenCalledWith({
        installedPackageRef: installedPackageReference,
      });
    });
  });
});

function setMockCoreClient(fnToMock: any, mockFn: jest.Mock<any, any>) {
  // Replace the specified function on the real KubeappsGrpcClient's
  // packages service implementation.
  const mockClient = new KubeappsGrpcClient().getPackagesServiceClientImpl();
  jest.spyOn(mockClient, fnToMock).mockImplementation(mockFn);
  jest.spyOn(InstalledPackage, "packagesServiceClient").mockImplementation(() => mockClient);
}

function setMockHelmClient(fnToMock: any, mockFn: jest.Mock<any, any>) {
  // Replace the specified function on the real KubeappsGrpcClient's
  // helm packages service implementation.
  const mockClient = new KubeappsGrpcClient().getHelmPackagesServiceClientImpl();
  jest.spyOn(mockClient, fnToMock).mockImplementation(mockFn);
  jest.spyOn(InstalledPackage, "helmPackagesServiceClient").mockImplementation(() => mockClient);
}
