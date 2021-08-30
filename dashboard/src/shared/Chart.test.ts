import {
  GetAvailablePackageDetailResponse,
  GetAvailablePackageSummariesResponse,
  GetAvailablePackageVersionsResponse,
  PackageAppVersion,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
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
        const availablePackageVersions = await Chart.getAvailablePackageVersions(
          t.args.cluster,
          t.args.namespace,
          t.args.id,
        );
        expect(availablePackageVersions).toStrictEqual({
          packageAppVersions: [
            { appVersion: "10.0.0", pkgVersion: "1.0.0" },
          ] as PackageAppVersion[],
        } as GetAvailablePackageVersionsResponse);
        expect(mockGetAvailablePackageVersions).toHaveBeenCalledWith(...Object.values(t.args));
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
          version: "v1",
        },
      },
      {
        description: "fetch availablePackageDetail latest version",
        args: {
          cluster: clusterName,
          namespace: namespaceName,
          id: "mypackage",
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
          t.args.cluster,
          t.args.namespace,
          t.args.id,
          t.args.version,
        );
        expect(availablePackageDetail).toStrictEqual({
          availablePackageDetail: { name: "foo" },
        } as GetAvailablePackageDetailResponse);
        expect(mockGetAvailablePackageDetail).toHaveBeenCalledWith(...Object.values(t.args));
      });
    });
  });
});
