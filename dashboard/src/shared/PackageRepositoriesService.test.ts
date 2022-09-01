// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import {
  AddPackageRepositoryResponse,
  DeletePackageRepositoryResponse,
  GetPackageRepositoryDetailResponse,
  GetPackageRepositorySummariesResponse,
  PackageRepositoryAuth_PackageRepositoryAuthType,
  PackageRepositoryReference,
  UpdatePackageRepositoryResponse,
} from "gen/kubeappsapis/core/packages/v1alpha1/repositories";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import { HelmPackageRepositoryCustomDetail } from "gen/kubeappsapis/plugins/helm/packages/v1alpha1/helm";
import { KappControllerPackageRepositoryCustomDetail } from "gen/kubeappsapis/plugins/kapp_controller/packages/v1alpha1/kapp_controller";
import KubeappsGrpcClient from "./KubeappsGrpcClient";
import { PackageRepositoriesService } from "./PackageRepositoriesService";
import { IPkgRepoFormData, PluginNames, RepositoryStorageTypes } from "./types";

const cluster = "cluster";
const namespace = "namespace";
const plugin: Plugin = { name: "my.plugin", version: "0.0.1" };

const helmCustomDetail: HelmPackageRepositoryCustomDetail = {
  imagesPullSecret: {
    secretRef: "test-1",
    credentials: undefined,
  },
  ociRepositories: ["apache", "jenkins"],
  performValidation: true,
  filterRule: {
    jq: ".name == $var0 or .name == $var1",
    variables: { $var0: "nginx", $var1: "wordpress" },
  },
};

const kappCustomDetail: KappControllerPackageRepositoryCustomDetail = {
  fetch: {
    imgpkgBundle: {
      tagSelection: {
        semver: {
          constraints: ">= 1.0.0",
          prereleases: {
            identifiers: ["alpha", "beta"],
          },
        },
      },
    },
    git: undefined,
    http: undefined,
    image: undefined,
    inline: undefined,
  },
};

const pkgRepoFormData = {
  plugin,
  authHeader: "",
  authMethod:
    PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED,
  basicAuth: {
    password: "",
    username: "",
  },
  customCA: "",
  customDetail: {},
  description: "",
  dockerRegCreds: {
    password: "",
    username: "",
    email: "",
    server: "",
  },
  interval: "",
  name: "",
  passCredentials: false,
  secretAuthName: "",
  secretTLSName: "",
  skipTLS: false,
  type: RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_HELM,
  url: "",
  opaqueCreds: {
    data: {},
  },
  sshCreds: {
    knownHosts: "",
    privateKey: "",
  },
  tlsCertKey: {
    cert: "",
    key: "",
  },
  namespace,
  isNamespaceScoped: true,
} as IPkgRepoFormData;

const packageRepoRef = {
  identifier: pkgRepoFormData.name,
  context: { cluster, namespace },
  plugin,
} as PackageRepositoryReference;

describe("RepositoriesService", () => {
  afterEach(() => {
    jest.restoreAllMocks();
  });

  it("getPackageRepositorySummaries", async () => {
    const mockGetPackageRepositorySummaries = jest.fn().mockImplementation(() =>
      Promise.resolve({
        packageRepositorySummaries: [
          { name: "repo1", packageRepoRef },
          { name: "repo2", packageRepoRef },
        ],
      } as GetPackageRepositorySummariesResponse),
    );
    setMockCoreClient("GetPackageRepositorySummaries", mockGetPackageRepositorySummaries);

    const getPackageRepositorySummariesResponse =
      await PackageRepositoriesService.getPackageRepositorySummaries({
        cluster,
        namespace,
      });
    expect(getPackageRepositorySummariesResponse).toStrictEqual({
      packageRepositorySummaries: [
        { name: "repo1", packageRepoRef },
        { name: "repo2", packageRepoRef },
      ],
    } as GetPackageRepositorySummariesResponse);
    expect(mockGetPackageRepositorySummaries).toHaveBeenCalledWith({
      context: { cluster, namespace },
    });
  });

  it("getPackageRepositoryDetail", async () => {
    const mockGetPackageRepositoryDetail = jest.fn().mockImplementation(() =>
      Promise.resolve({
        detail: { name: "repo1", packageRepoRef },
      } as GetPackageRepositoryDetailResponse),
    );
    setMockCoreClient("GetPackageRepositoryDetail", mockGetPackageRepositoryDetail);

    const getPackageRepositoryDetailResponse =
      await PackageRepositoriesService.getPackageRepositoryDetail(packageRepoRef);
    expect(getPackageRepositoryDetailResponse).toStrictEqual({
      detail: { name: "repo1", packageRepoRef },
    } as GetPackageRepositoryDetailResponse);
    expect(mockGetPackageRepositoryDetail).toHaveBeenCalledWith({ packageRepoRef });
  });

  it("addPackageRepository", async () => {
    const mockAddPackageRepository = jest.fn().mockImplementation(() =>
      Promise.resolve({
        packageRepoRef: {
          identifier: pkgRepoFormData.name,
          context: { cluster, namespace },
          plugin,
        } as PackageRepositoryReference,
      } as AddPackageRepositoryResponse),
    );
    setMockCoreClient("AddPackageRepository", mockAddPackageRepository);

    const addPackageRepositoryResponse = await PackageRepositoriesService.addPackageRepository(
      cluster,
      { ...pkgRepoFormData, namespace, isNamespaceScoped: false },
    );
    expect(addPackageRepositoryResponse).toStrictEqual({
      packageRepoRef: {
        identifier: pkgRepoFormData.name,
        context: { cluster, namespace },
        plugin,
      } as PackageRepositoryReference,
    } as AddPackageRepositoryResponse);
    expect(mockAddPackageRepository).toHaveBeenCalledWith({
      context: { cluster, namespace },
      description: "",
      interval: "",
      name: "",
      namespaceScoped: false,
      plugin,
      type: "helm",
      url: "",
    });
  });

  it("updatePackageRepository", async () => {
    const mockUpdatePackageRepository = jest.fn().mockImplementation(() =>
      Promise.resolve({
        packageRepoRef: {
          identifier: pkgRepoFormData.name,
          context: { cluster, namespace },
          plugin,
        } as PackageRepositoryReference,
      } as UpdatePackageRepositoryResponse),
    );
    setMockCoreClient("UpdatePackageRepository", mockUpdatePackageRepository);

    const updatePackageRepositoryResponse =
      await PackageRepositoriesService.updatePackageRepository(cluster, pkgRepoFormData);
    expect(updatePackageRepositoryResponse).toStrictEqual({
      packageRepoRef: {
        identifier: pkgRepoFormData.name,
        context: { cluster, namespace },
        plugin,
      } as PackageRepositoryReference,
    } as UpdatePackageRepositoryResponse);
    expect(mockUpdatePackageRepository).toHaveBeenCalledWith({
      description: "",
      interval: "",
      packageRepoRef: {
        context: {
          cluster: cluster,
          namespace: namespace,
        },
        identifier: "",
        plugin,
      },
      url: "",
    });
  });

  it("deletePackageRepository", async () => {
    const mockDeletePackageRepository = jest.fn().mockImplementation(() =>
      Promise.resolve({
        packageRepoRef: {
          identifier: pkgRepoFormData.name,
          context: { cluster, namespace },
          plugin,
        } as PackageRepositoryReference,
      } as DeletePackageRepositoryResponse),
    );
    setMockCoreClient("DeletePackageRepository", mockDeletePackageRepository);

    const deletePackageRepositoryResponse =
      await PackageRepositoriesService.deletePackageRepository({
        identifier: pkgRepoFormData.name,
        context: { cluster, namespace },
        plugin,
      } as PackageRepositoryReference);
    expect(deletePackageRepositoryResponse).toStrictEqual({
      packageRepoRef: {
        identifier: pkgRepoFormData.name,
        context: { cluster, namespace },
        plugin,
      } as PackageRepositoryReference,
    } as DeletePackageRepositoryResponse);
    expect(mockDeletePackageRepository).toHaveBeenCalledWith({
      packageRepoRef: {
        context: { cluster, namespace },
        identifier: "",
        plugin,
      },
    });
  });
});

describe("buildEncodedCustomDetail encoding", () => {
  it("returns undefined if the plugin is not supported)", async () => {
    const encodedCustomDetail = PackageRepositoriesService["buildEncodedCustomDetail"]({
      ...pkgRepoFormData,
      plugin: { name: "my.plugin", version: "0.0.1" },
      customDetail: undefined,
    });
    expect(encodedCustomDetail).toStrictEqual(undefined);
  });

  it("returns undefined if no custom details (helm)", async () => {
    const encodedCustomDetail = PackageRepositoriesService["buildEncodedCustomDetail"]({
      ...pkgRepoFormData,
      plugin: { name: PluginNames.PACKAGES_HELM, version: "v1alpha1" },
      customDetail: undefined,
    });
    expect(encodedCustomDetail).toBe(undefined);
  });

  it("returns encoded empty value if no custom details (kapp)", async () => {
    const encodedCustomDetail = PackageRepositoriesService["buildEncodedCustomDetail"]({
      ...pkgRepoFormData,
      plugin: { name: PluginNames.PACKAGES_KAPP, version: "v1alpha1" },
      customDetail: undefined,
    });
    expect(encodedCustomDetail).toBe(undefined);
  });

  it("encodes the custom details (helm)", async () => {
    const encodedCustomDetail = PackageRepositoriesService["buildEncodedCustomDetail"]({
      ...pkgRepoFormData,
      plugin: { name: PluginNames.PACKAGES_HELM, version: "v1alpha1" },
      customDetail: helmCustomDetail,
    });
    expect(encodedCustomDetail?.typeUrl).toBe(
      "kubeappsapis.plugins.helm.packages.v1alpha1.HelmPackageRepositoryCustomDetail",
    );
    expect(encodedCustomDetail?.value.byteLength).toBe(101);
    expect(
      HelmPackageRepositoryCustomDetail.decode(encodedCustomDetail?.value as any),
    ).toStrictEqual(helmCustomDetail);
  });

  it("encodes the custom details (kapp)", async () => {
    const encodedCustomDetail = PackageRepositoriesService["buildEncodedCustomDetail"]({
      ...pkgRepoFormData,
      plugin: { name: PluginNames.PACKAGES_KAPP, version: "v1alpha1" },
      customDetail: kappCustomDetail,
    });
    expect(encodedCustomDetail?.typeUrl).toBe(
      "kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackageRepositoryCustomDetail",
    );
    expect(encodedCustomDetail?.value.byteLength).toBe(33);
    expect(
      KappControllerPackageRepositoryCustomDetail.decode(encodedCustomDetail?.value as any),
    ).toStrictEqual(kappCustomDetail);
  });
});

function setMockCoreClient(fnToMock: any, mockFn: jest.Mock<any, any>) {
  // Replace the specified function on the real KubeappsGrpcClient's
  // packages service implementation.
  const mockClient = new KubeappsGrpcClient().getRepositoriesServiceClientImpl();
  jest.spyOn(mockClient, fnToMock).mockImplementation(mockFn);
  jest
    .spyOn(PackageRepositoriesService, "coreRepositoriesClient")
    .mockImplementation(() => mockClient);
}
