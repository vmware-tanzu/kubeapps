// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import {
  AvailablePackageReference,
  InstalledPackageDetail,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import {
  AddPackageRepositoryResponse,
  DeletePackageRepositoryResponse,
  GetPackageRepositoryDetailResponse,
  GetPackageRepositorySummariesResponse,
  PackageRepositoryAuth_PackageRepositoryAuthType,
  PackageRepositoryDetail,
  PackageRepositoryReference,
  PackageRepositorySummary,
  UpdatePackageRepositoryResponse,
} from "gen/kubeappsapis/core/packages/v1alpha1/repositories";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import { HelmPackageRepositoryCustomDetail } from "gen/kubeappsapis/plugins/helm/packages/v1alpha1/helm";
import context from "jest-plugin-context";
import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";
import { PackageRepositoriesService } from "shared/PackageRepositoriesService";
import PackagesService from "shared/PackagesService";
import { initialState } from "shared/specs/mountWrapper";
import {
  IPkgRepoFormData,
  IStoreState,
  NotFoundNetworkError,
  PluginNames,
  RepositoryStorageTypes,
} from "shared/types";
import { getType } from "typesafe-actions";
import actions from ".";
import { convertPkgRepoDetailToSummary } from "./repos";

const { repos: repoActions } = actions;
const mockStore = configureMockStore([thunk]);

let store: any;
const plugin = { name: PluginNames.PACKAGES_HELM, version: "0.0.1" } as Plugin;
const fluxPlugin = { name: PluginNames.PACKAGES_FLUX, version: "v1beta1" } as Plugin;
const carvelPlugin = { name: PluginNames.PACKAGES_KAPP, version: "v1beta1" } as Plugin;

const packageRepoRef = {
  identifier: "repo-abc",
  context: { cluster: "default", namespace: "default" },
  plugin: plugin,
} as PackageRepositoryReference;

const packageRepositorySummary = {
  name: "repo-abc",
  description: "",
  namespaceScoped: false,
  type: "helm",
  url: "https://helm.repo",
  packageRepoRef: packageRepoRef,
} as PackageRepositorySummary;

const packageRepositoryDetail = {
  name: "repo-abc",
  type: "helm",
  description: "",
  interval: "10m",
  namespaceScoped: false,
  url: "https://helm.repo",
  packageRepoRef: {
    identifier: "repo-abc",
    context: { cluster: "default", namespace: "default" },
    plugin: plugin,
  },
} as PackageRepositoryDetail;

const kubeappsNamespace = "kubeapps-namespace";
const helmGlobalNamespace = "kubeapps-repos-global";
const carvelGlobalNamespace = "carvel-repos-global";

beforeEach(() => {
  store = mockStore({
    config: {
      ...initialState.config,
      kubeappsNamespace,
      helmGlobalNamespace,
      carvelGlobalNamespace,
    },
    clusters: {
      ...initialState.clusters,
      currentCluster: "default",
      clusters: {
        ...initialState.clusters.clusters,
        default: {
          ...initialState.clusters.clusters[initialState.clusters.currentCluster],
          currentNamespace: kubeappsNamespace,
        },
      },
    },
  } as Partial<IStoreState>);

  PackageRepositoriesService.getPackageRepositorySummaries = jest
    .fn()
    .mockImplementationOnce(() => {
      return {
        packageRepositorySummaries: [packageRepositorySummary],
      } as GetPackageRepositorySummariesResponse;
    });
  PackageRepositoriesService.deletePackageRepository = jest.fn().mockImplementationOnce(() => {
    return {} as DeletePackageRepositoryResponse;
  });
  PackageRepositoriesService.getPackageRepositoryDetail = jest.fn().mockImplementationOnce(() => {
    return { detail: packageRepositoryDetail } as GetPackageRepositoryDetailResponse;
  });
  PackageRepositoriesService.updatePackageRepository = jest.fn().mockImplementationOnce(() => {
    return {
      packageRepoRef: packageRepoRef,
    } as UpdatePackageRepositoryResponse;
  });
  PackageRepositoriesService.addPackageRepository = jest.fn().mockImplementationOnce(() => {
    return {
      packageRepoRef: packageRepoRef,
    } as AddPackageRepositoryResponse;
  });
});

afterEach(jest.restoreAllMocks);

// Regular action creators
interface ITestCase {
  name: string;
  action: (...args: any[]) => any;
  args?: any;
  payload?: any;
}

const pkgRepoFormData = {
  plugin: plugin,
  authHeader: "",
  authMethod:
    PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED,
  basicAuth: {
    password: "",
    username: "",
  },
  customCA: "",
  customDetail: {
    imagesPullSecret: {
      secretRef: "repo-1",
      credentials: { server: "", username: "", password: "", email: "" },
    },
    ociRepositories: [],
    performValidation: false,
    filterRules: [],
  } as HelmPackageRepositoryCustomDetail,
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
  namespace: "my-namespace",
  isNamespaceScoped: true,
} as IPkgRepoFormData;

const actionTestCases: ITestCase[] = [
  { name: "addOrUpdateRepo", action: repoActions.addOrUpdateRepo },
  {
    name: "addedRepo",
    action: repoActions.addedRepo,
    args: packageRepositoryDetail,
    payload: packageRepositoryDetail,
  },
  { name: "requestRepoSummaries", action: repoActions.requestRepoSummaries },
  {
    name: "receiveRepoSummaries",
    action: repoActions.receiveRepoSummaries,
    args: [[packageRepositorySummary]],
    payload: [packageRepositorySummary],
  },
  { name: "requestRepoDetail", action: repoActions.requestRepoDetail },
  {
    name: "receiveRepoDetail",
    action: repoActions.receiveRepoDetail,
    args: [packageRepositoryDetail],
    payload: packageRepositoryDetail,
  },
  {
    name: "errorRepos",
    action: repoActions.errorRepos,
    args: [new Error("foo"), "create"],
    payload: { err: new Error("foo"), op: "create" },
  },
];

actionTestCases.forEach(tc => {
  describe(tc.name, () => {
    it("has expected structure", () => {
      const actionResult =
        tc.args && tc.args.length && typeof tc.args === "object"
          ? tc.action.call(null, ...tc.args)
          : tc.action.call(null, tc.args);
      expect(actionResult).toEqual({
        type: getType(tc.action),
        payload: tc.payload,
      });
    });
  });
});

// Async action creators
describe("deleteRepo", () => {
  context("dispatches requestRepoSummaries and receivedRepos after deletion if no error", () => {
    const currentNamespace = "current-namespace";
    it("dispatches requestRepoSummaries with current namespace", async () => {
      const storeWithFlag: any = mockStore({
        clusters: {
          ...initialState.clusters,
          currentCluster: "defaultCluster",
          clusters: {
            ...initialState.clusters.clusters,
            defaultCluster: {
              ...initialState.clusters.clusters[initialState.clusters.currentCluster],
              currentNamespace,
            },
          },
        },
      } as Partial<IStoreState>);
      await storeWithFlag.dispatch(
        repoActions.deleteRepo({
          context: { cluster: "default", namespace: "my-namespace" },
          identifier: "foo",
          plugin: plugin,
        } as PackageRepositoryReference),
      );
      expect(storeWithFlag.getActions()).toEqual([]);
    });
  });

  it("dispatches errorRepos if error deleting", async () => {
    PackageRepositoriesService.deletePackageRepository = jest.fn().mockImplementationOnce(() => {
      throw new Error("Boom!");
    });

    const expectedActions = [
      {
        type: getType(repoActions.errorRepos),
        payload: { err: new Error("Boom!"), op: "delete" },
      },
    ];

    await store.dispatch(
      repoActions.deleteRepo({
        context: { cluster: "default", namespace: "my-namespace" },
        identifier: "foo",
        plugin: plugin,
      } as PackageRepositoryReference),
    );
    expect(store.getActions()).toEqual(expectedActions);
  });
});

describe("fetchRepoSummaries", () => {
  const namespace = "default";
  it("dispatches requestRepoSummaries and receivedRepos if no error", async () => {
    const expectedActions = [
      {
        type: getType(repoActions.requestRepoSummaries),
        payload: namespace,
      },
      {
        type: getType(repoActions.receiveRepoSummaries),
        payload: [packageRepositorySummary],
      },
    ];

    await store.dispatch(repoActions.fetchRepoSummaries(namespace));
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("dispatches requestRepoSummaries and errorRepos if error fetching", async () => {
    PackageRepositoriesService.getPackageRepositorySummaries = jest
      .fn()
      .mockImplementationOnce(() => {
        throw new Error("Boom!");
      });

    const expectedActions = [
      {
        type: getType(repoActions.requestRepoSummaries),
        payload: namespace,
      },
      {
        type: getType(repoActions.errorRepos),
        payload: { err: new Error("Boom!"), op: "fetch" },
      },
    ];

    await store.dispatch(repoActions.fetchRepoSummaries(namespace));
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("fetches additional repos from the global namespace and joins them", async () => {
    PackageRepositoriesService.getPackageRepositorySummaries = jest
      .fn()
      .mockImplementationOnce(() => {
        return {
          packageRepositorySummaries: [{ name: "repo1", packageRepoRef: { identifier: "repo1" } }],
        } as GetPackageRepositorySummariesResponse;
      })
      .mockImplementationOnce(() => {
        return {
          packageRepositorySummaries: [{ name: "repo2", packageRepoRef: { identifier: "repo2" } }],
        } as GetPackageRepositorySummariesResponse;
      });

    const expectedActions = [
      {
        type: getType(repoActions.requestRepoSummaries),
        payload: namespace,
      },
      {
        type: getType(repoActions.requestRepoSummaries),
        payload: "",
      },
      {
        type: getType(repoActions.receiveRepoSummaries),
        payload: [
          { name: "repo1", packageRepoRef: { identifier: "repo1" } },
          { name: "repo2", packageRepoRef: { identifier: "repo2" } },
        ] as PackageRepositorySummary[],
      },
    ];
    await store.dispatch(repoActions.fetchRepoSummaries(namespace, true));
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("fetches duplicated repos from several namespaces and joins them", async () => {
    PackageRepositoriesService.getPackageRepositorySummaries = jest
      .fn()
      .mockImplementationOnce(() => {
        return {
          packageRepositorySummaries: [{ name: "repo1", packageRepoRef: { identifier: "repo1" } }],
        } as GetPackageRepositorySummariesResponse;
      })
      .mockImplementationOnce(() => {
        return {
          packageRepositorySummaries: [
            { name: "repo1", packageRepoRef: { identifier: "repo1" } },
            { name: "repo2", packageRepoRef: { identifier: "repo2" } },
          ],
        } as GetPackageRepositorySummariesResponse;
      });

    const expectedActions = [
      {
        type: getType(repoActions.requestRepoSummaries),
        payload: namespace,
      },
      {
        type: getType(repoActions.requestRepoSummaries),
        payload: "",
      },
      {
        type: getType(repoActions.receiveRepoSummaries),
        payload: [
          { name: "repo1", packageRepoRef: { identifier: "repo1" } },
          { name: "repo2", packageRepoRef: { identifier: "repo2" } },
        ] as PackageRepositorySummary[],
      },
    ];

    await store.dispatch(repoActions.fetchRepoSummaries(namespace, true));
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("fetches repos only if the namespace is the one used for global repos", async () => {
    PackageRepositoriesService.getPackageRepositorySummaries = jest
      .fn()
      .mockImplementationOnce(() => {
        return {
          packageRepositorySummaries: [{ name: "repo1" }],
        } as GetPackageRepositorySummariesResponse;
      })
      .mockImplementationOnce(() => {
        return {
          packageRepositorySummaries: [{ name: "repo1" }, { name: "repo2" }],
        } as GetPackageRepositorySummariesResponse;
      });

    const expectedActions = [
      {
        type: getType(repoActions.requestRepoSummaries),
        payload: helmGlobalNamespace,
      },
      {
        type: getType(repoActions.receiveRepoSummaries),
        payload: [{ name: "repo1" }],
      },
    ];

    await store.dispatch(repoActions.fetchRepoSummaries(helmGlobalNamespace, true));
    expect(store.getActions()).toEqual(expectedActions);
  });
});

describe("addRepo", () => {
  const addRepoCMD = repoActions.addRepo(pkgRepoFormData);

  context("when authHeader provided", () => {
    const addRepoCMDAuth = repoActions.addRepo({
      ...pkgRepoFormData,
      authHeader: "Bearer: abc",
    });

    it("calls PackageRepositoriesService create including a auth struct (authHeader)", async () => {
      await store.dispatch(addRepoCMDAuth);
      expect(PackageRepositoriesService.addPackageRepository).toHaveBeenCalledWith("default", {
        ...pkgRepoFormData,
        authHeader: "Bearer: abc",
      });
    });

    it("calls PackageRepositoriesService create including ociRepositories", async () => {
      await store.dispatch(
        repoActions.addRepo({
          ...pkgRepoFormData,
          type: RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_OCI,
          customDetail: {
            ...pkgRepoFormData.customDetail,
            ociRepositories: ["apache", "jenkins"],
          },
        }),
      );
      expect(PackageRepositoriesService.addPackageRepository).toHaveBeenCalledWith("default", {
        ...pkgRepoFormData,

        type: RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_OCI,
        customDetail: {
          ...pkgRepoFormData.customDetail,
          ociRepositories: ["apache", "jenkins"],
        },
      });
    });

    it("calls PackageRepositoriesService create skipping TLS verification", async () => {
      await store.dispatch(repoActions.addRepo({ ...pkgRepoFormData, skipTLS: true }));
      expect(PackageRepositoriesService.addPackageRepository).toHaveBeenCalledWith("default", {
        ...pkgRepoFormData,
        skipTLS: true,
      });
    });

    it("returns true", async () => {
      const res = await store.dispatch(addRepoCMDAuth);
      expect(res).toBe(true);
    });
  });

  context("when a customCA is provided", () => {
    const addRepoCMDAuth = repoActions.addRepo({
      ...pkgRepoFormData,
      customCA: "This is a cert!",
    });

    it("calls PackageRepositoriesService create including a auth struct (custom CA)", async () => {
      await store.dispatch(addRepoCMDAuth);
      expect(PackageRepositoriesService.addPackageRepository).toHaveBeenCalledWith("default", {
        ...pkgRepoFormData,
        customCA: "This is a cert!",
      });
    });

    it("returns true (addRepoCMDAuth)", async () => {
      const res = await store.dispatch(addRepoCMDAuth);
      expect(res).toBe(true);
    });

    it("sets flux repos as global", async () => {
      await store.dispatch(
        repoActions.addRepo({
          ...pkgRepoFormData,
          namespace: "my-namespace",
          isNamespaceScoped: false,
          plugin: fluxPlugin as Plugin,
        }),
      );
      expect(PackageRepositoriesService.addPackageRepository).toHaveBeenCalledWith("default", {
        ...pkgRepoFormData,
        namespace: "my-namespace",
        isNamespaceScoped: false,
        plugin: fluxPlugin,
      });
    });

    it("sets carvel repos as global if using the carvelGlobalNamespace", async () => {
      await store.dispatch(
        repoActions.addRepo({
          ...pkgRepoFormData,
          namespace: carvelGlobalNamespace,
          isNamespaceScoped: false,
          plugin: carvelPlugin as Plugin,
        }),
      );
      expect(PackageRepositoriesService.addPackageRepository).toHaveBeenCalledWith("default", {
        ...pkgRepoFormData,
        namespace: carvelGlobalNamespace,
        isNamespaceScoped: false,
        plugin: carvelPlugin,
      });
    });
  });

  context("when authHeader and customCA are empty", () => {
    it("calls PackageRepositoriesService create without a auth struct", async () => {
      await store.dispatch(addRepoCMD);
      expect(PackageRepositoriesService.addPackageRepository).toHaveBeenCalledWith(
        "default",
        pkgRepoFormData,
      );
    });

    it("returns true (addRepoCMD)", async () => {
      const res = await store.dispatch(addRepoCMD);
      expect(res).toBe(true);
    });
  });

  it("dispatches addOrUpdateRepo and errorRepos if error fetching", async () => {
    PackageRepositoriesService.addPackageRepository = jest.fn().mockImplementationOnce(() => {
      throw new Error("Boom!");
    });

    const expectedActions = [
      {
        type: getType(repoActions.addOrUpdateRepo),
      },
      {
        type: getType(repoActions.errorRepos),
        payload: { err: new Error("Boom!"), op: "create" },
      },
    ];

    await store.dispatch(addRepoCMD);
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("returns false if error fetching", async () => {
    PackageRepositoriesService.addPackageRepository = jest.fn().mockImplementationOnce(() => {
      throw new Error("Boom!");
    });

    const res = await store.dispatch(addRepoCMD);
    expect(res).toEqual(false);
  });

  it("dispatches addOrUpdateRepo and addedRepo if no error", async () => {
    const expectedActions = [
      {
        type: getType(repoActions.addOrUpdateRepo),
      },
      {
        type: getType(repoActions.addedRepo),
        payload: convertPkgRepoDetailToSummary(packageRepositoryDetail),
      },
    ];

    await store.dispatch(addRepoCMD);
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("includes registry secrets if given", async () => {
    await store.dispatch(
      repoActions.addRepo({
        ...pkgRepoFormData,
        customDetail: {
          ...pkgRepoFormData.customDetail,
          imagesPullSecret: {
            secretRef: "repo-1",
            credentials: { server: "", username: "", password: "", email: "" },
          },
        },
      }),
    );

    expect(PackageRepositoriesService.addPackageRepository).toHaveBeenCalledWith("default", {
      ...pkgRepoFormData,
      customDetail: {
        ...pkgRepoFormData.customDetail,
        imagesPullSecret: {
          secretRef: "repo-1",
          credentials: { server: "", username: "", password: "", email: "" },
        },
      },
    });
  });

  it("calls PackageRepositoriesService create with description", async () => {
    await store.dispatch(
      repoActions.addRepo({
        ...pkgRepoFormData,
        description: "This is a weird description 123!@#$%^&&*()_+-=<>?/.,;:'\"",
      }),
    );
    expect(PackageRepositoriesService.addPackageRepository).toHaveBeenCalledWith("default", {
      ...pkgRepoFormData,
      description: "This is a weird description 123!@#$%^&&*()_+-=<>?/.,;:'\"",
    });
  });
});

describe("updateRepo", () => {
  it("updates a repo with an auth header", async () => {
    const pkgRepoDetail = {
      ...packageRepositoryDetail,
      auth: {
        header: "foo",
      },
    } as PackageRepositoryDetail;

    PackageRepositoriesService.updatePackageRepository = jest.fn().mockReturnValue({
      packageRepoRef: pkgRepoDetail.packageRepoRef,
    } as UpdatePackageRepositoryResponse);
    PackageRepositoriesService.getPackageRepositoryDetail = jest.fn().mockReturnValue({
      detail: pkgRepoDetail,
    } as GetPackageRepositoryDetailResponse);
    const expectedActions = [
      {
        type: getType(repoActions.addOrUpdateRepo),
      },
      {
        type: getType(repoActions.repoUpdated),
        payload: convertPkgRepoDetailToSummary(pkgRepoDetail),
      },
    ];

    await store.dispatch(
      repoActions.updateRepo({
        ...pkgRepoFormData,
        authHeader: "foo",
      }),
    );

    expect(store.getActions()).toEqual(expectedActions);
    expect(PackageRepositoriesService.updatePackageRepository).toHaveBeenCalledWith("default", {
      ...pkgRepoFormData,
      authHeader: "foo",
    });
  });

  it("updates a repo with an customCA", async () => {
    const pkgRepoDetail = {
      ...packageRepositoryDetail,
      tlsConfig: {
        secretRef: { name: "pkgrepo-repo-abc", key: "data" },
        certAuthority: "",
        insecureSkipVerify: false,
      },
    } as PackageRepositoryDetail;
    PackageRepositoriesService.updatePackageRepository = jest.fn().mockReturnValue({
      packageRepoRef: packageRepositoryDetail.packageRepoRef,
    } as UpdatePackageRepositoryResponse);
    PackageRepositoriesService.getPackageRepositoryDetail = jest.fn().mockReturnValue({
      detail: pkgRepoDetail,
    } as GetPackageRepositoryDetailResponse);
    const expectedActions = [
      {
        type: getType(repoActions.addOrUpdateRepo),
      },
      {
        type: getType(repoActions.repoUpdated),
        payload: convertPkgRepoDetailToSummary(pkgRepoDetail),
      },
    ];

    await store.dispatch(
      repoActions.updateRepo({
        ...pkgRepoFormData,
        customCA: "This is a cert!",
      }),
    );
    expect(store.getActions()).toEqual(expectedActions);
    expect(PackageRepositoriesService.updatePackageRepository).toHaveBeenCalledWith("default", {
      ...pkgRepoFormData,
      customCA: "This is a cert!",
    });
  });

  it("returns an error if failed", async () => {
    PackageRepositoriesService.updatePackageRepository = jest.fn(() => {
      throw new Error("boom");
    });
    const expectedActions = [
      {
        type: getType(repoActions.addOrUpdateRepo),
      },
      {
        type: getType(repoActions.errorRepos),
        payload: { err: new Error("boom"), op: "update" },
      },
    ];

    await store.dispatch(repoActions.updateRepo(pkgRepoFormData));
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("updates a repo with ociRepositories", async () => {
    PackageRepositoriesService.updatePackageRepository = jest.fn().mockReturnValue({
      packageRepoRef: packageRepositoryDetail.packageRepoRef,
    } as UpdatePackageRepositoryResponse);
    await store.dispatch(
      repoActions.updateRepo({
        ...pkgRepoFormData,
        type: RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_OCI,
        customDetail: { ...pkgRepoFormData.customDetail, ociRepositories: ["apache", "jenkins"] },
      }),
    );
    expect(PackageRepositoriesService.updatePackageRepository).toHaveBeenCalledWith("default", {
      ...pkgRepoFormData,
      type: RepositoryStorageTypes.PACKAGE_REPOSITORY_STORAGE_OCI,
      customDetail: { ...pkgRepoFormData.customDetail, ociRepositories: ["apache", "jenkins"] },
    });
  });

  it("updates a repo with description", async () => {
    PackageRepositoriesService.updatePackageRepository = jest.fn().mockReturnValue({
      packageRepoRef: packageRepositoryDetail.packageRepoRef,
    } as UpdatePackageRepositoryResponse);
    await store.dispatch(
      repoActions.updateRepo({
        ...pkgRepoFormData,
        description: "updated description",
      }),
    );
    expect(PackageRepositoriesService.updatePackageRepository).toHaveBeenCalledWith("default", {
      ...pkgRepoFormData,
      description: "updated description",
    });
  });
});

describe("findPackageInRepo", () => {
  const installedPackageDetail = {
    availablePackageRef: {
      context: { cluster: "default", namespace: "my-ns" },
      identifier: "my-repo/my-package",
      plugin: plugin,
    },
  } as InstalledPackageDetail;
  it("dispatches requestRepo and receivedRepo if no error", async () => {
    PackagesService.getAvailablePackageVersions = jest.fn();
    const expectedActions = [
      {
        type: getType(repoActions.requestRepoDetail),
      },
      {
        type: getType(repoActions.receiveRepoDetail),
        payload: packageRepositoryDetail,
      },
    ];
    await store.dispatch(
      repoActions.findPackageInRepo(
        "default",
        "other-namespace",
        "my-repo",
        installedPackageDetail,
      ),
    );
    expect(store.getActions()).toEqual(expectedActions);
    expect(PackagesService.getAvailablePackageVersions).toBeCalledWith({
      context: { cluster: "default", namespace: "other-namespace" },
      identifier: "my-repo/my-package",
      plugin: plugin,
    } as AvailablePackageReference);
  });

  it("dispatches requestRepo and createErrorPackage if error fetching", async () => {
    PackagesService.getAvailablePackageVersions = jest.fn(() => {
      throw new Error();
    });

    const expectedActions = [
      {
        type: getType(repoActions.requestRepoDetail),
      },
      {
        type: getType(repoActions.errorRepos),
        payload: {
          err: new NotFoundNetworkError(
            "Package my-repo/my-package not found in the repository other-namespace.",
          ),
          op: "fetch",
        },
      },
    ];

    await store.dispatch(
      repoActions.findPackageInRepo(
        "default",
        "other-namespace",
        "my-repo",
        installedPackageDetail,
      ),
    );
    expect(store.getActions()).toEqual(expectedActions);
    expect(PackagesService.getAvailablePackageVersions).toBeCalledWith({
      context: { cluster: "default", namespace: "other-namespace" },
      identifier: "my-repo/my-package",
      plugin: plugin,
    } as AvailablePackageReference);
  });
});
