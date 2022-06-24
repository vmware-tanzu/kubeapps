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
import KubeappsGrpcClient from "./KubeappsGrpcClient";
import { PackageRepositoriesService } from "./PackageRepositoriesService";
import { IPkgRepoFormData, RepositoryStorageTypes } from "./types";

const cluster = "cluster";
const namespace = "namespace";
const plugin: Plugin = { name: "my.plugin", version: "0.0.1" };

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
  customDetails: {
    dockerRegistrySecrets: [],
    ociRepositories: [],
    performValidation: false,
    filterRules: [],
  },
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
      namespace,
      pkgRepoFormData,
      false,
    );
    expect(addPackageRepositoryResponse).toStrictEqual({
      packageRepoRef: {
        identifier: pkgRepoFormData.name,
        context: { cluster, namespace },
        plugin,
      } as PackageRepositoryReference,
    } as AddPackageRepositoryResponse);
    expect(mockAddPackageRepository).toHaveBeenCalledWith({
      auth: { opaqueCreds: { data: {} } },
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
      await PackageRepositoriesService.updatePackageRepository(cluster, namespace, pkgRepoFormData);
    expect(updatePackageRepositoryResponse).toStrictEqual({
      packageRepoRef: {
        identifier: pkgRepoFormData.name,
        context: { cluster, namespace },
        plugin,
      } as PackageRepositoryReference,
    } as UpdatePackageRepositoryResponse);
    expect(mockUpdatePackageRepository).toHaveBeenCalledWith({
      auth: {
        opaqueCreds: {
          //TODO(agamez): check this
          data: {},
        },
      },
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

function setMockCoreClient(fnToMock: any, mockFn: jest.Mock<any, any>) {
  // Replace the specified function on the real KubeappsGrpcClient's
  // packages service implementation.
  const mockClient = new KubeappsGrpcClient().getRepositoriesServiceClientImpl();
  jest.spyOn(mockClient, fnToMock).mockImplementation(mockFn);
  jest
    .spyOn(PackageRepositoriesService, "coreRepositoriesClient")
    .mockImplementation(() => mockClient);
}
