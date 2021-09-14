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
import Chart from "./Chart";

const clusterName = "cluster-name";
const namespaceName = "namespace-name";
const defaultPage = 1;
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
          cluster: clusterName,
          namespace: namespaceName,
          repos: "",
          page: defaultPage,
          size: defaultSize,
          query: "",
        },
      },
      {
        description: "fetch availablePackageSummaries without repos, with query",
        args: {
          cluster: clusterName,
          namespace: namespaceName,
          repos: "",
          page: defaultPage,
          size: defaultSize,
          query: "cms",
        },
      },
      {
        description: "fetch availablePackageSummaries with repos, without query",
        args: {
          cluster: clusterName,
          namespace: namespaceName,
          repos: "repo1,repo2",
          page: defaultPage,
          size: defaultSize,
          query: "",
        },
      },
      {
        description: "fetch availablePackageSummaries with repos, with query",
        args: {
          cluster: clusterName,
          namespace: namespaceName,
          repos: "repo1,repo2",
          page: defaultPage,
          size: defaultSize,
          query: "cms",
        },
      },
    ].forEach(t => {
      it(t.description, async () => {
        const mockGetAvailablePackageSummaries = jest.fn().mockImplementation(() =>
          Promise.resolve({
            availablePackageSummaries: [{ name: "foo" }],
            nextPageToken: "",
            categories: ["foo"],
          } as GetAvailablePackageSummariesResponse),
        );
        jest
          .spyOn(Chart, "getAvailablePackageSummaries")
          .mockImplementation(mockGetAvailablePackageSummaries);
        const availablePackageSummaries = await Chart.getAvailablePackageSummaries(
          t.args.cluster,
          t.args.namespace,
          t.args.repos,
          t.args.page,
          t.args.size,
          t.args.query,
        );
        expect(availablePackageSummaries).toStrictEqual({
          availablePackageSummaries: [{ name: "foo" }],
          nextPageToken: "",
          categories: ["foo"],
        } as GetAvailablePackageSummariesResponse);
        expect(mockGetAvailablePackageSummaries).toHaveBeenCalledWith(...Object.values(t.args));
      });
    });
  });
  describe("getAvailablePackageVersions", () => {
    [
      {
        description: "fetch availablePackageVersions",
        args: {
          cluster: clusterName,
          namespace: namespaceName,
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
          .spyOn(Chart, "getAvailablePackageVersions")
          .mockImplementation(mockGetAvailablePackageVersions);
        const availablePackageVersions = await Chart.getAvailablePackageVersions({
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
          cluster: clusterName,
          namespace: namespaceName,
          id: "mypackage",
          plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
          version: "v1",
        },
      },
      {
        description: "fetch availablePackageDetail latest version",
        args: {
          cluster: clusterName,
          namespace: namespaceName,
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
          .spyOn(Chart, "getAvailablePackageDetail")
          .mockImplementation(mockGetAvailablePackageDetail);
        const availablePackageDetail = await Chart.getAvailablePackageDetail(
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
