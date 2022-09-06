// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import {
  AvailablePackageReference,
  GetAvailablePackageDetailResponse,
  GetAvailablePackageSummariesResponse,
  GetAvailablePackageVersionsResponse,
  PackageAppVersion,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import * as moxios from "moxios";
import { axiosWithAuth } from "./AxiosInstance";
import { KubeappsGrpcClient } from "./KubeappsGrpcClient";
import PackagesService from "./PackagesService";

const cluster = "cluster-name";
const namespace = "namespace-name";
const defaultPageToken = "defaultPageToken";
const defaultSize = 0;
describe("App", () => {
  beforeEach(() => {
    // Import as "any" to avoid typescript syntax error
    moxios.install(axiosWithAuth as any);
    moxios.stubRequest(/.*/, {
      response: { data: "ok" },
      status: 200,
    });
  });
  afterEach(() => {
    moxios.uninstall(axiosWithAuth as any);
    jest.restoreAllMocks();
  });
  describe("getAvailablePackageSummaries", () => {
    [
      {
        description: "fetch availablePackageSummaries without repos, without query",
        args: {
          cluster: cluster,
          namespace: namespace,
          repos: "",
          nextPageToken: defaultPageToken,
          size: defaultSize,
          query: "",
        },
        expectedClientArg: {
          context: { cluster, namespace },
          filterOptions: {
            query: "",
            repositories: [],
          },
          paginationOptions: { pageToken: defaultPageToken, pageSize: defaultSize },
        },
      },
      {
        description: "fetch availablePackageSummaries without repos, with query",
        args: {
          cluster: cluster,
          namespace: namespace,
          repos: "",
          nextPageToken: defaultPageToken,
          size: defaultSize,
          query: "cms",
        },
        expectedClientArg: {
          context: { cluster, namespace },
          filterOptions: {
            query: "cms",
            repositories: [],
          },
          paginationOptions: { pageToken: defaultPageToken, pageSize: defaultSize },
        },
      },
      {
        description: "fetch availablePackageSummaries with repos, without query",
        args: {
          cluster: cluster,
          namespace: namespace,
          repos: "repo1,repo2",
          nextPageToken: defaultPageToken,
          size: defaultSize,
          query: "",
        },
        expectedClientArg: {
          context: { cluster, namespace },
          filterOptions: {
            query: "",
            repositories: ["repo1", "repo2"],
          },
          paginationOptions: { pageToken: defaultPageToken, pageSize: defaultSize },
        },
      },
      {
        description: "fetch availablePackageSummaries with repos, with query",
        args: {
          cluster: cluster,
          namespace: namespace,
          repos: "repo1,repo2",
          nextPageToken: defaultPageToken,
          size: defaultSize,
          query: "cms",
        },
        expectedClientArg: {
          context: { cluster, namespace },
          filterOptions: {
            query: "cms",
            repositories: ["repo1", "repo2"],
          },
          paginationOptions: { pageToken: defaultPageToken, pageSize: defaultSize },
        },
      },
    ].forEach(t => {
      it(t.description, async () => {
        const mockClientGetAvailablePackageSummaries = jest.fn().mockImplementation(() =>
          Promise.resolve({
            availablePackageSummaries: [{ name: "foo" }],
            nextPageToken: "",
            categories: ["foo"],
          } as GetAvailablePackageSummariesResponse),
        );
        // Create a real client, but we'll stub out the function we're interested in.
        const mockClient = new KubeappsGrpcClient().getPackagesServiceClientImpl();
        jest
          .spyOn(mockClient, "GetAvailablePackageSummaries")
          .mockImplementation(mockClientGetAvailablePackageSummaries);
        jest.spyOn(PackagesService, "packagesServiceClient").mockImplementation(() => mockClient);
        const availablePackageSummaries = await PackagesService.getAvailablePackageSummaries(
          t.args.cluster,
          t.args.namespace,
          t.args.repos,
          t.args.nextPageToken,
          t.args.size,
          t.args.query,
        );
        expect(availablePackageSummaries).toStrictEqual({
          availablePackageSummaries: [{ name: "foo" }],
          nextPageToken: "",
          categories: ["foo"],
        } as GetAvailablePackageSummariesResponse);
        expect(mockClientGetAvailablePackageSummaries).toHaveBeenCalledWith(t.expectedClientArg);
      });
    });
  });
  describe("getAvailablePackageVersions", () => {
    [
      {
        description: "fetch availablePackageVersions",
        args: {
          cluster: cluster,
          namespace: namespace,
          id: "mypackage",
          plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
        },
      },
    ].forEach(t => {
      it(t.description, async () => {
        const mockGetAvailablePackageVersions = jest.fn().mockImplementation(() =>
          Promise.resolve({
            packageAppVersions: [
              { appVersion: "10.0.0", pkgVersion: "1.0.0" },
            ] as PackageAppVersion[],
          } as GetAvailablePackageVersionsResponse),
        );
        jest
          .spyOn(PackagesService, "getAvailablePackageVersions")
          .mockImplementation(mockGetAvailablePackageVersions);
        const availablePackageVersions = await PackagesService.getAvailablePackageVersions({
          context: { cluster: t.args.cluster, namespace: t.args.namespace },
          identifier: t.args.id,
          plugin: t.args.plugin,
        } as AvailablePackageReference);
        expect(availablePackageVersions).toStrictEqual({
          packageAppVersions: [
            { appVersion: "10.0.0", pkgVersion: "1.0.0" },
          ] as PackageAppVersion[],
        } as GetAvailablePackageVersionsResponse);
        expect(mockGetAvailablePackageVersions).toHaveBeenCalledWith({
          context: { cluster: t.args.cluster, namespace: t.args.namespace },
          identifier: t.args.id,
          plugin: t.args.plugin,
        } as AvailablePackageReference);
      });
    });
  });
  describe("getAvailablePackageDetail", () => {
    [
      {
        description: "fetch availablePackageDetail with version",
        args: {
          cluster: cluster,
          namespace: namespace,
          id: "mypackage",
          plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
          version: "v1",
        },
      },
      {
        description: "fetch availablePackageDetail latest version",
        args: {
          cluster: cluster,
          namespace: namespace,
          id: "mypackage",
          plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
          version: undefined,
        },
      },
    ].forEach(t => {
      it(t.description, async () => {
        const mockGetAvailablePackageDetail = jest.fn().mockImplementation(() =>
          Promise.resolve({
            availablePackageDetail: { name: "foo" },
          } as GetAvailablePackageDetailResponse),
        );
        jest
          .spyOn(PackagesService, "getAvailablePackageDetail")
          .mockImplementation(mockGetAvailablePackageDetail);
        const availablePackageDetail = await PackagesService.getAvailablePackageDetail(
          {
            context: { cluster: t.args.cluster, namespace: t.args.namespace },
            identifier: t.args.id,
            plugin: t.args.plugin,
          } as AvailablePackageReference,
          t.args.version,
        );
        expect(availablePackageDetail).toStrictEqual({
          availablePackageDetail: { name: "foo" },
        } as GetAvailablePackageDetailResponse);
        expect(mockGetAvailablePackageDetail).toHaveBeenCalledWith(
          {
            context: { cluster: t.args.cluster, namespace: t.args.namespace },
            identifier: t.args.id,
            plugin: t.args.plugin,
          } as AvailablePackageReference,
          t.args.version,
        );
      });
    });
  });
});
