// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import {
  AvailablePackageReference,
  InstalledPackageDetail,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import {
  AddPackageRepositoryResponse,
  PackageRepositoryAuth_PackageRepositoryAuthType,
  PackageRepositoryDetail,
  PackageRepositoryReference,
  PackageRepositorySummary,
} from "gen/kubeappsapis/core/packages/v1alpha1/repositories";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import context from "jest-plugin-context";
import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";
import { PackageRepositoriesService } from "shared/PackageRepositoriesService";
import PackagesService from "shared/PackagesService";
import Secret from "shared/Secret";
import { NotFoundError } from "shared/types";
import { getType } from "typesafe-actions";
import actions from ".";

const { repos: repoActions } = actions;
const mockStore = configureMockStore([thunk]);

let store: any;
const appRepo = { name: "repo-abc" } as PackageRepositoryDetail;
const kubeappsNamespace = "kubeapps-namespace";
const globalReposNamespace = "kubeapps-repos-global";

const plugin: Plugin = { name: "my.plugin", version: "0.0.1" };

beforeEach(() => {
  store = mockStore({
    config: { kubeappsNamespace, globalReposNamespace },
    clusters: {
      currentCluster: "default",
      clusters: {
        default: {
          currentNamespace: kubeappsNamespace,
        },
      },
    },
  });
  PackageRepositoriesService.getPackageRepositorySummaries = jest
    .fn()
    .mockImplementationOnce(() => {
      return { items: { foo: "bar" } };
    });
  PackageRepositoriesService.deletePackageRepository = jest.fn();
  PackageRepositoriesService.getPackageRepositoryDetail = jest.fn().mockImplementationOnce(() => {
    return appRepo;
  });
  PackageRepositoriesService.updatePackageRepository = jest.fn();
  PackageRepositoriesService.addPackageRepository = jest.fn().mockImplementationOnce(() => {
    return {
      packageRepoRef: {
        identifier: "repo-abc",
        context: { cluster: "", namespace: "" },
        plugin: { name: "", version: "" },
      },
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

const repo = { name: "my-repo" } as PackageRepositorySummary;

const actionTestCases: ITestCase[] = [
  { name: "addRepo", action: repoActions.addRepo },
  { name: "addedRepo", action: repoActions.addedRepo, args: repo, payload: repo },
  { name: "requestRepos", action: repoActions.requestRepos },
  { name: "receiveRepos", action: repoActions.receiveRepos, args: [[repo]], payload: [repo] },
  { name: "requestRepo", action: repoActions.requestRepo },
  { name: "receiveRepo", action: repoActions.receiveRepo, args: repo, payload: repo },
  { name: "redirect", action: repoActions.redirect, args: "/foo", payload: "/foo" },
  { name: "redirected", action: repoActions.redirected },
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
  context("dispatches requestRepos and receivedRepos after deletion if no error", () => {
    const currentNamespace = "current-namespace";
    it("dispatches requestRepos with current namespace", async () => {
      const storeWithFlag: any = mockStore({
        clusters: {
          currentCluster: "defaultCluster",
          clusters: {
            defaultCluster: {
              currentNamespace,
            },
          },
        },
      });
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

describe("fetchRepos", () => {
  const namespace = "default";
  it("dispatches requestRepos and receivedRepos if no error", async () => {
    const expectedActions = [
      {
        type: getType(repoActions.requestRepos),
        payload: namespace,
      },
      {
        type: getType(repoActions.receiveRepos),
        payload: { foo: "bar" },
      },
    ];

    await store.dispatch(repoActions.fetchRepos(namespace));
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("dispatches requestRepos and errorRepos if error fetching", async () => {
    PackageRepositoriesService.getPackageRepositorySummaries = jest
      .fn()
      .mockImplementationOnce(() => {
        throw new Error("Boom!");
      });

    const expectedActions = [
      {
        type: getType(repoActions.requestRepos),
        payload: namespace,
      },
      {
        type: getType(repoActions.errorRepos),
        payload: { err: new Error("Boom!"), op: "fetch" },
      },
    ];

    await store.dispatch(repoActions.fetchRepos(namespace));
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("fetches additional repos from the global namespace and joins them", async () => {
    PackageRepositoriesService.getPackageRepositorySummaries = jest
      .fn()
      .mockImplementationOnce(() => {
        return { items: [{ name: "repo1", metadata: { uid: "123" } }] };
      })
      .mockImplementationOnce(() => {
        return { items: [{ name: "repo2", metadata: { uid: "321" } }] };
      });

    const expectedActions = [
      {
        type: getType(repoActions.requestRepos),
        payload: namespace,
      },
      {
        type: getType(repoActions.requestRepos),
        payload: globalReposNamespace,
      },
      {
        type: getType(repoActions.receiveRepos),
        payload: [
          { name: "repo1", metadata: { uid: "123" } },
          { name: "repo2", metadata: { uid: "321" } },
        ],
      },
    ];

    await store.dispatch(repoActions.fetchRepos(namespace, true));
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("fetches duplicated repos from several namespaces and joins them", async () => {
    PackageRepositoriesService.getPackageRepositorySummaries = jest
      .fn()
      .mockImplementationOnce(() => {
        return { items: [{ name: "repo1", metadata: { uid: "123" } }] };
      })
      .mockImplementationOnce(() => {
        return {
          items: [
            { name: "repo2", metadata: { uid: "321" } },
            { name: "repo3", metadata: { uid: "321" } },
          ],
        };
      });

    const expectedActions = [
      {
        type: getType(repoActions.requestRepos),
        payload: namespace,
      },
      {
        type: getType(repoActions.requestRepos),
        payload: globalReposNamespace,
      },
      {
        type: getType(repoActions.receiveRepos),
        payload: [
          { name: "repo1", metadata: { uid: "123" } },
          { name: "repo2", metadata: { uid: "321" } },
        ],
      },
    ];

    await store.dispatch(repoActions.fetchRepos(namespace, true));
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("fetches repos only if the namespace is the one used for global repos", async () => {
    PackageRepositoriesService.getPackageRepositorySummaries = jest
      .fn()
      .mockImplementationOnce(() => {
        return { items: [{ name: "repo1", metadata: { uid: "123" } }] };
      })
      .mockImplementationOnce(() => {
        return {
          items: [
            { name: "repo1", metadata: { uid: "321" } },
            { name: "repo2", metadata: { uid: "123" } },
          ],
        };
      });

    const expectedActions = [
      {
        type: getType(repoActions.requestRepos),
        payload: globalReposNamespace,
      },
      {
        type: getType(repoActions.receiveRepos),
        payload: [{ name: "repo1", metadata: { uid: "123" } }],
      },
    ];

    await store.dispatch(repoActions.fetchRepos(globalReposNamespace, true));
    expect(store.getActions()).toEqual(expectedActions);
  });
});

describe("installRepo", () => {
  const installRepoCMD = repoActions.installRepo(
    "my-repo",
    plugin,
    "my-namespace",
    "http://foo.bar",
    "helm",
    "",
    "",
    "",
    "",
    [],
    [],
    false,
    false,
    PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED,
    3600,
    "user",
    "password",
    true,
    undefined,
  );

  context("when authHeader provided", () => {
    const installRepoCMDAuth = repoActions.installRepo(
      "my-repo",
      plugin,
      "my-namespace",
      "http://foo.bar",
      "helm",
      "",
      "Bearer: abc",
      "",
      "",
      [],
      [],
      false,
      false,
      PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED,
      3600,
      "user",
      "password",
      true,
      undefined,
    );

    it("calls AppRepository create including a auth struct (authHeader)", async () => {
      await store.dispatch(installRepoCMDAuth);
      expect(PackageRepositoriesService.addPackageRepository).toHaveBeenCalledWith(
        "default",
        "my-repo",
        "my-namespace",
        "http://foo.bar",
        "helm",
        "",
        "Bearer: abc",
        "",
        "",
        {},
        [],
        [],
        false,
        false,
        undefined,
      );
    });

    it("calls AppRepository create including ociRepositories", async () => {
      await store.dispatch(
        repoActions.installRepo(
          "my-repo",
          plugin,
          "my-namespace",
          "http://foo.bar",
          "oci",
          "",
          "",
          "",
          "",
          [],
          ["apache", "jenkins"],
          false,
          false,
          PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED,
          3600,
          "user",
          "password",
          true,
          undefined,
        ),
      );
      expect(PackageRepositoriesService.addPackageRepository).toHaveBeenCalledWith(
        "default",
        "my-repo",
        "my-namespace",
        "http://foo.bar",
        "oci",
        "",
        "",
        "",
        "",
        {},
        [],
        ["apache", "jenkins"],
        false,
        false,
        undefined,
      );
    });

    it("calls AppRepository create skipping TLS verification", async () => {
      await store.dispatch(
        repoActions.installRepo(
          "my-repo",
          plugin,
          "my-namespace",
          "http://foo.bar",
          "oci",
          "",
          "",
          "",
          "",
          [],
          [],
          true,
          false,
          PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED,
          3600,
          "user",
          "password",
          true,
          undefined,
        ),
      );
      expect(PackageRepositoriesService.addPackageRepository).toHaveBeenCalledWith(
        "default",
        "my-repo",
        "my-namespace",
        "http://foo.bar",
        "oci",
        "",
        "",
        "",
        "",
        {},
        [],
        [],
        true,
        false,
        undefined,
      );
    });

    it("returns true", async () => {
      const res = await store.dispatch(installRepoCMDAuth);
      expect(res).toBe(true);
    });
  });

  context("when a customCA is provided", () => {
    const installRepoCMDAuth = repoActions.installRepo(
      "my-repo",
      plugin,
      "my-namespace",
      "http://foo.bar",
      "helm",
      "",
      "",
      "",
      "This is a cert!",
      [],
      [],
      false,
      false,
      PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED,
      3600,
      "user",
      "password",
      true,
      undefined,
    );

    it("calls AppRepository create including a auth struct (custom CA)", async () => {
      await store.dispatch(installRepoCMDAuth);
      expect(PackageRepositoriesService.addPackageRepository).toHaveBeenCalledWith(
        "default",
        "my-repo",
        "my-namespace",
        "http://foo.bar",
        "helm",
        "",
        "",
        "",
        "This is a cert!",
        {},
        [],
        [],
        false,
        false,
        undefined,
      );
    });

    it("returns true (installRepoCMDAuth)", async () => {
      const res = await store.dispatch(installRepoCMDAuth);
      expect(res).toBe(true);
    });

    context("when a pod template is provided", () => {
      it("calls AppRepository create including pod template", async () => {
        await store.dispatch(
          repoActions.installRepo(
            "my-repo",
            plugin,
            "my-namespace",
            "http://foo.bar",
            "helm",
            "",
            "",
            "",
            "",
            [],
            [],
            false,
            false,
            PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED,
            3600,
            "user",
            "password",
            true,
            undefined,
          ),
        );

        expect(PackageRepositoriesService.addPackageRepository).toHaveBeenCalledWith(
          "default",
          plugin,
          "my-repo",
          "my-namespace",
          "http://foo.bar",
          "helm",
          "",
          "",
          "",
          "",
          {
            spec: { containers: [{ env: [{ name: "FOO", value: "BAR" }] }] },
          },
          [],
          [],
          false,
          false,
          undefined,
        );
      });
    });
  });

  context("when authHeader and customCA are empty", () => {
    it("calls AppRepository create without a auth struct", async () => {
      await store.dispatch(installRepoCMD);
      expect(PackageRepositoriesService.addPackageRepository).toHaveBeenCalledWith(
        "default",
        "my-repo",
        "my-namespace",
        "http://foo.bar",
        "helm",
        "",
        "",
        "",
        "",
        {},
        [],
        [],
        false,
        false,
        undefined,
      );
    });

    it("returns true (installRepoCMD)", async () => {
      const res = await store.dispatch(installRepoCMD);
      expect(res).toBe(true);
    });
  });

  it("dispatches addRepo and errorRepos if error fetching", async () => {
    PackageRepositoriesService.addPackageRepository = jest.fn().mockImplementationOnce(() => {
      throw new Error("Boom!");
    });

    const expectedActions = [
      {
        type: getType(repoActions.addRepo),
      },
      {
        type: getType(repoActions.errorRepos),
        payload: { err: new Error("Boom!"), op: "create" },
      },
    ];

    await store.dispatch(installRepoCMD);
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("returns false if error fetching", async () => {
    PackageRepositoriesService.addPackageRepository = jest.fn().mockImplementationOnce(() => {
      throw new Error("Boom!");
    });

    const res = await store.dispatch(installRepoCMD);
    expect(res).toEqual(false);
  });

  it("dispatches addRepo and addedRepo if no error", async () => {
    const expectedActions = [
      {
        type: getType(repoActions.addRepo),
      },
      {
        type: getType(repoActions.addedRepo),
        payload: { metadata: { name: "repo-abc" } },
      },
    ];

    await store.dispatch(installRepoCMD);
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("includes registry secrets if given", async () => {
    await store.dispatch(
      repoActions.installRepo(
        "my-repo",
        plugin,
        "foo",
        "http://foo.bar",
        "helm",
        "",
        "",
        "",
        "",
        ["repo-1"],
        [],
        false,
        false,
        PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED,
        3600,
        "user",
        "password",
        true,
        undefined,
      ),
    );

    expect(PackageRepositoriesService.addPackageRepository).toHaveBeenCalledWith(
      "default",
      "my-repo",
      "foo",
      "http://foo.bar",
      "helm",
      "",
      "",
      "",
      "",
      {},
      ["repo-1"],
      [],
      false,
      false,
      PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED,
      undefined,
    );
  });

  it("calls AppRepository create with description", async () => {
    await store.dispatch(
      repoActions.installRepo(
        "my-repo",
        plugin,
        "my-namespace",
        "http://foo.bar",
        "oci",
        "This is a weird description 123!@#$%^&&*()_+-=<>?/.,;:'\"",
        "",
        "",
        "",
        [],
        ["apache", "jenkins"],
        false,
        false,
        PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED,
        3600,
        "user",
        "password",
        true,
        undefined,
      ),
    );
    expect(PackageRepositoriesService.addPackageRepository).toHaveBeenCalledWith(
      "default",
      "my-repo",
      "my-namespace",
      "http://foo.bar",
      "oci",
      "This is a weird description 123!@#$%^&&*()_+-=<>?/.,;:'\"",
      "",
      "",
      "",
      {},
      [],
      ["apache", "jenkins"],
      false,
      false,
      PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED,
      undefined,
    );
  });
});

describe("updateRepo", () => {
  it("updates a repo with an auth header", async () => {
    const r = {
      metadata: { name: "repo-abc" },
      spec: { auth: { header: { secretKeyRef: { name: "apprepo-repo-abc" } } } },
    };
    PackageRepositoriesService.updatePackageRepository = jest.fn().mockReturnValue({
      appRepository: r,
    });
    const expectedActions = [
      {
        type: getType(repoActions.requestRepoUpdate),
      },
      {
        type: getType(repoActions.repoUpdated),
        payload: r,
      },
    ];

    await store.dispatch(
      repoActions.updateRepo(
        "my-repo",
        plugin,
        "my-namespace",
        "http://foo.bar",
        "helm",
        "",
        "foo",
        "",
        "bar",
        ["repo-1"],
        [],
        false,
        false,
        PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED,
        3600,
        "user",
        "password",
        true,
        undefined,
      ),
    );
    expect(store.getActions()).toEqual(expectedActions);
    expect(PackageRepositoriesService.updatePackageRepository).toHaveBeenCalledWith(
      "default",
      "my-repo",
      "my-namespace",
      "http://foo.bar",
      "helm",
      "",
      "foo",
      "",
      "bar",
      { spec: { containers: [{ env: [{ name: "FOO", value: "BAR" }] }] } },
      ["repo-1"],
      [],
      false,
      false,
      PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED,
      undefined,
    );
  });

  it("updates a repo with an customCA", async () => {
    const r = {
      metadata: { name: "repo-abc" },
      spec: { auth: { customCA: { secretKeyRef: { name: "apprepo-repo-abc" } } } },
    };
    PackageRepositoriesService.updatePackageRepository = jest.fn().mockReturnValue({
      appRepository: r,
    });
    const expectedActions = [
      {
        type: getType(repoActions.requestRepoUpdate),
      },
      {
        type: getType(repoActions.repoUpdated),
        payload: r,
      },
    ];

    await store.dispatch(
      repoActions.updateRepo(
        "my-repo",
        plugin,
        "my-namespace",
        "http://foo.bar",
        "helm",
        "",
        "foo",
        "",
        "bar",
        ["repo-1"],
        [],
        false,
        false,
        PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED,
        3600,
        "user",
        "password",
        true,
        undefined,
      ),
    );
    expect(store.getActions()).toEqual(expectedActions);
    expect(PackageRepositoriesService.updatePackageRepository).toHaveBeenCalledWith(
      "default",
      "my-repo",
      "my-namespace",
      "http://foo.bar",
      "helm",
      "",
      "foo",
      "",
      "bar",
      { spec: { containers: [{ env: [{ name: "FOO", value: "BAR" }] }] } },
      ["repo-1"],
      [],
      false,
      false,
      PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED,
      undefined,
    );
  });

  it("returns an error if failed", async () => {
    PackageRepositoriesService.updatePackageRepository = jest.fn(() => {
      throw new Error("boom");
    });
    const expectedActions = [
      {
        type: getType(repoActions.requestRepoUpdate),
      },
      {
        type: getType(repoActions.errorRepos),
        payload: { err: new Error("boom"), op: "update" },
      },
    ];

    await store.dispatch(
      repoActions.updateRepo(
        "my-repo",
        plugin,
        "my-namespace",
        "http://foo.bar",
        "helm",
        "",
        "foo",
        "",
        "bar",
        [],
        [],
        false,
        false,
        PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED,
        3600,
        "user",
        "password",
        true,
        undefined,
      ),
    );
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("updates a repo with ociRepositories", async () => {
    PackageRepositoriesService.updatePackageRepository = jest.fn().mockReturnValue({
      appRepository: {},
    });
    await store.dispatch(
      repoActions.updateRepo(
        "my-repo",
        plugin,
        "my-namespace",
        "http://foo.bar",
        "oci",
        "",
        "",
        "",
        "",
        [],
        ["apache", "jenkins"],
        false,
        false,
        PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED,
        3600,
        "user",
        "password",
        true,
        undefined,
      ),
    );
    expect(PackageRepositoriesService.updatePackageRepository).toHaveBeenCalledWith(
      "default",
      "my-repo",
      "my-namespace",
      "http://foo.bar",
      "oci",
      "",
      "",
      "",
      "",
      {},
      [],
      ["apache", "jenkins"],
      false,
      false,
      undefined,
    );
  });

  it("updates a repo with description", async () => {
    PackageRepositoriesService.updatePackageRepository = jest.fn().mockReturnValue({
      appRepository: {},
    });
    await store.dispatch(
      repoActions.updateRepo(
        "my-repo",
        plugin,
        "my-namespace",
        "http://foo.bar",
        "oci",
        "updated description",
        "",
        "",
        "",
        [],
        ["apache", "jenkins"],
        false,
        false,
        PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED,
        3600,
        "user",
        "password",
        true,
        undefined,
      ),
    );
    expect(PackageRepositoriesService.updatePackageRepository).toHaveBeenCalledWith(
      "default",
      "my-repo",
      "my-namespace",
      "http://foo.bar",
      "oci",
      "updated description",
      "",
      "",
      "",
      {},
      [],
      ["apache", "jenkins"],
      false,
      false,
      PackageRepositoryAuth_PackageRepositoryAuthType.PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED,
      undefined,
    );
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
        type: getType(repoActions.requestRepo),
      },
      {
        type: getType(repoActions.receiveRepo),
        payload: appRepo,
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
        type: getType(repoActions.requestRepo),
      },
      {
        type: getType(actions.availablepackages.createErrorPackage),
        payload: new NotFoundError(
          "Package my-repo/my-package not found in the repository other-namespace.",
        ),
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

describe("createDockerRegistrySecret", () => {
  it("creates a docker registry", async () => {
    Secret.createPullSecret = jest.fn();
    const expectedActions = [
      {
        type: getType(repoActions.createImagePullSecret),
        payload: "secret-name",
      },
    ];

    await store.dispatch(
      repoActions.createDockerRegistrySecret(
        "secret-name",
        "user",
        "password",
        "email",
        "server",
        "namespace",
      ),
    );
    expect(Secret.createPullSecret).toHaveBeenCalledWith(
      "default",
      "secret-name",
      "user",
      "password",
      "email",
      "server",
      "namespace",
    );
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("dispatches an error", async () => {
    Secret.createPullSecret = jest.fn(() => {
      throw new Error("boom");
    });
    const expectedActions = [
      {
        type: getType(repoActions.errorRepos),
        payload: {
          err: new Error("boom"),
          op: "fetch",
        },
      },
    ];
    await store.dispatch(repoActions.createDockerRegistrySecret("", "", "", "", "", ""));
    expect(store.getActions()).toEqual(expectedActions);
  });
});
