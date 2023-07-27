// Copyright 2022-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import {
  AddPackageRepositoryResponse,
  DeletePackageRepositoryResponse,
  GetPackageRepositoryDetailResponse,
  GetPackageRepositoryPermissionsResponse,
  GetPackageRepositorySummariesResponse,
  PackageRepositoriesPermissions,
  PackageRepositoryAuth_PackageRepositoryAuthType,
  PackageRepositoryReference,
  UpdatePackageRepositoryResponse,
} from "gen/kubeappsapis/core/packages/v1alpha1/repositories_pb";
import { Context } from "gen/kubeappsapis/core/packages/v1alpha1/packages_pb";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins_pb";
import {
  HelmPackageRepositoryCustomDetail,
  ImagesPullSecret,
  PodSecurityContext,
  ProxyOptions,
  RepositoryFilterRule,
} from "gen/kubeappsapis/plugins/helm/packages/v1alpha1/helm_pb";
import {
  KappControllerPackageRepositoryCustomDetail,
  PackageRepositoryFetch,
  PackageRepositoryImgpkg,
  VersionSelection,
  VersionSelectionSemver,
  VersionSelectionSemverPrereleases,
} from "gen/kubeappsapis/plugins/kapp_controller/packages/v1alpha1/kapp_controller_pb";
import KubeappsGrpcClient from "./KubeappsGrpcClient";
import { PackageRepositoriesService } from "./PackageRepositoriesService";
import { IPkgRepoFormData, PluginNames, RepositoryStorageTypes } from "./types";

const cluster = "cluster";
const namespace = "namespace";
const plugin: Plugin = new Plugin({ name: "my.plugin", version: "0.0.1" });

const helmCustomDetail: HelmPackageRepositoryCustomDetail = new HelmPackageRepositoryCustomDetail({
  imagesPullSecret: new ImagesPullSecret({
    dockerRegistryCredentialOneOf: {
      case: "secretRef",
      value: "test-1",
    },
  }),
  ociRepositories: ["apache", "jenkins"],
  performValidation: true,
  filterRule: new RepositoryFilterRule({
    jq: ".name == $var0 or .name == $var1",
    variables: { $var0: "nginx", $var1: "wordpress" },
  }),
  proxyOptions: new ProxyOptions({
    enabled: true,
    httpProxy: "http://proxy",
    httpsProxy: "https://proxy",
    noProxy: "localhost",
  }),
  // these options are not used by the UI
  tolerations: [],
  nodeSelector: {},
  securityContext: new PodSecurityContext({
    supplementalGroups: [],
    fSGroup: undefined,
    runAsGroup: undefined,
    runAsNonRoot: undefined,
    runAsUser: undefined,
  }),
});

const kappCustomDetail = new KappControllerPackageRepositoryCustomDetail({
  fetch: new PackageRepositoryFetch({
    imgpkgBundle: new PackageRepositoryImgpkg({
      tagSelection: new VersionSelection({
        semver: new VersionSelectionSemver({
          constraints: ">= 1.0.0",
          prereleases: new VersionSelectionSemverPrereleases({
            identifiers: ["alpha", "beta"],
          }),
        }),
      }),
    }),
    git: undefined,
    http: undefined,
    image: undefined,
    inline: undefined,
  }),
});

const pkgRepoFormData = {
  plugin,
  authHeader: "",
  authMethod: PackageRepositoryAuth_PackageRepositoryAuthType.UNSPECIFIED,
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
    setMockCoreClient("getPackageRepositorySummaries", mockGetPackageRepositorySummaries);

    const getPackageRepositorySummariesResponse =
      await PackageRepositoriesService.getPackageRepositorySummaries(
        new Context({
          cluster,
          namespace,
        }),
      );
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
    setMockCoreClient("getPackageRepositoryDetail", mockGetPackageRepositoryDetail);

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
    setMockCoreClient("addPackageRepository", mockAddPackageRepository);

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
    setMockCoreClient("updatePackageRepository", mockUpdatePackageRepository);

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
      Promise.resolve(
        new DeletePackageRepositoryResponse({
          packageRepoRef: {
            identifier: pkgRepoFormData.name,
            context: { cluster, namespace },
            plugin,
          } as PackageRepositoryReference,
        }),
      ),
    );
    setMockCoreClient("deletePackageRepository", mockDeletePackageRepository);

    const deletePackageRepositoryResponse =
      await PackageRepositoriesService.deletePackageRepository({
        identifier: pkgRepoFormData.name,
        context: { cluster, namespace },
        plugin,
      } as PackageRepositoryReference);
    expect(deletePackageRepositoryResponse).toStrictEqual(
      new DeletePackageRepositoryResponse({
        packageRepoRef: new PackageRepositoryReference({
          identifier: pkgRepoFormData.name,
          context: { cluster, namespace },
          plugin,
        }),
      }),
    );
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
      plugin: new Plugin({ name: "my.plugin", version: "0.0.1" }),
      customDetail: undefined,
    });
    expect(encodedCustomDetail).toStrictEqual(undefined);
  });

  it("returns undefined if no custom details (helm)", async () => {
    const encodedCustomDetail = PackageRepositoriesService["buildEncodedCustomDetail"]({
      ...pkgRepoFormData,
      plugin: new Plugin({ name: PluginNames.PACKAGES_HELM, version: "v1alpha1" }),
      customDetail: undefined,
    });
    expect(encodedCustomDetail).toBe(undefined);
  });

  it("returns encoded empty value if no custom details (kapp)", async () => {
    const encodedCustomDetail = PackageRepositoriesService["buildEncodedCustomDetail"]({
      ...pkgRepoFormData,
      plugin: new Plugin({ name: PluginNames.PACKAGES_KAPP, version: "v1alpha1" }),
      customDetail: undefined,
    });
    expect(encodedCustomDetail).toBe(undefined);
  });

  it("encodes the custom details (helm)", async () => {
    const encodedCustomDetail = PackageRepositoriesService["buildEncodedCustomDetail"]({
      ...pkgRepoFormData,
      plugin: new Plugin({ name: PluginNames.PACKAGES_HELM, version: "v1alpha1" }),
      customDetail: helmCustomDetail,
    });
    expect(encodedCustomDetail?.typeUrl).toBe(
      "kubeappsapis.plugins.helm.packages.v1alpha1.HelmPackageRepositoryCustomDetail",
    );
    expect(encodedCustomDetail?.value.byteLength).toBe(147);
    expect(HelmPackageRepositoryCustomDetail.fromBinary(encodedCustomDetail!.value)).toStrictEqual(
      helmCustomDetail,
    );
  });

  it("encodes the custom details (kapp)", async () => {
    const encodedCustomDetail = PackageRepositoriesService["buildEncodedCustomDetail"]({
      ...pkgRepoFormData,
      plugin: new Plugin({ name: PluginNames.PACKAGES_KAPP, version: "v1alpha1" }),
      customDetail: kappCustomDetail,
    });
    expect(encodedCustomDetail?.typeUrl).toBe(
      "kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackageRepositoryCustomDetail",
    );
    expect(encodedCustomDetail?.value).toStrictEqual(kappCustomDetail);
  });

  it("getRepositoriesPermissions", async () => {
    const mockGetRepositoriesPermissions = jest.fn().mockImplementation(() =>
      Promise.resolve({
        permissions: [
          new PackageRepositoriesPermissions({
            plugin: plugin,
            global: {
              create: true,
            },
            namespace: {
              list: true,
            },
          }),
        ],
      } as GetPackageRepositoryPermissionsResponse),
    );
    setMockCoreClient("getPackageRepositoryPermissions", mockGetRepositoriesPermissions);

    const getPackageRepositoryPermissionsResponse =
      await PackageRepositoriesService.getRepositoriesPermissions(cluster, namespace);
    expect(getPackageRepositoryPermissionsResponse).toStrictEqual([
      new PackageRepositoriesPermissions({
        plugin: plugin,
        global: {
          create: true,
        },
        namespace: {
          list: true,
        },
      }),
    ] as PackageRepositoriesPermissions[]);
    expect(mockGetRepositoriesPermissions).toHaveBeenCalledWith({
      context: { cluster, namespace },
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
