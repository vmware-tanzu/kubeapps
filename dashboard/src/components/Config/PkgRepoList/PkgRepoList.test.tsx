// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import actions from "actions";
import Alert from "components/js/Alert";
import Table from "components/js/Table";
import TableRow from "components/js/Table/components/TableRow";
import Tooltip from "components/js/Tooltip";
import { PackageRepositorySummary } from "gen/kubeappsapis/core/packages/v1alpha1/repositories";
import { act } from "react-dom/test-utils";
import * as ReactRedux from "react-redux";
import { Link } from "react-router-dom";
import { IPackageRepositoryState } from "reducers/repos";
import { Kube } from "shared/Kube";
import { defaultStore, getStore, initialState, mountWrapper } from "shared/specs/mountWrapper";
import { IStoreState } from "shared/types";
import { PkgRepoControl } from "./PkgRepoControl";
import { PkgRepoDisabledControl } from "./PkgRepoDisabledControl";
import PkgRepoList from "./PkgRepoList";

const {
  clusters: { currentCluster, clusters },
  config: { helmGlobalNamespace },
} = initialState;
const namespace = clusters[currentCluster].currentNamespace;

let spyOnUseDispatch: jest.SpyInstance;
const kubeaActions = { ...actions.kube };
beforeEach(() => {
  actions.repos = {
    ...actions.repos,
    fetchRepoSummaries: jest.fn(),
  };
  const mockDispatch = jest.fn();
  spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
  Kube.canI = jest.fn().mockReturnValue({
    then: jest.fn(f => f(true)),
    catch: jest.fn(f => f(false)),
  });
});

afterEach(() => {
  actions.kube = { ...kubeaActions };
  spyOnUseDispatch.mockRestore();
});

it("fetches repos and imagePullSecrets", () => {
  mountWrapper(defaultStore, <PkgRepoList />);
  expect(actions.repos.fetchRepoSummaries).toHaveBeenCalledWith(namespace, true);
});

it("fetches repos only from the helmGlobalNamespace", () => {
  mountWrapper(
    getStore({
      clusters: {
        ...initialState.clusters,
        clusters: {
          [currentCluster]: {
            ...initialState.clusters.clusters[currentCluster],
            currentNamespace: helmGlobalNamespace,
          },
        },
      },
    } as Partial<IStoreState>),
    <PkgRepoList />,
  );
  expect(actions.repos.fetchRepoSummaries).toHaveBeenCalledWith("");
});

it("fetches repos from all namespaces (without kubeappsNamespace)", () => {
  const wrapper = mountWrapper(defaultStore, <PkgRepoList />);
  act(() => {
    wrapper.find("input[type='checkbox']").simulate("change");
  });
  expect(actions.repos.fetchRepoSummaries).toHaveBeenCalledWith("");
});

it("should hide the all-namespace switch if the user doesn't have permissions", async () => {
  Kube.canI = jest.fn().mockReturnValue({
    then: jest.fn((f: any) => f(false)),
    catch: jest.fn(f => f(false)),
  });
  const wrapper = mountWrapper(defaultStore, <PkgRepoList />);
  expect(wrapper.find("input[type='checkbox']")).not.toExist();
});

// TODO: Remove this test when package repos are supported in different clusters
it("shows a warning if the cluster is not the default one", () => {
  const wrapper = mountWrapper(
    getStore({
      clusters: {
        ...initialState.clusters,
        currentCluster: "other",
        clusters: {
          ...initialState.clusters.clusters,
          other: {
            ...initialState.clusters.clusters[currentCluster],
          },
        },
      },
    } as Partial<IStoreState>),
    <PkgRepoList />,
  );
  expect(wrapper.find(Alert)).toIncludeText(
    "Package Repositories can't be managed from this cluster",
  );
});

it("shows an error fetching a repo", () => {
  const wrapper = mountWrapper(
    getStore({
      repos: { errors: { fetch: new Error("boom!") } } as IPackageRepositoryState,
    } as Partial<IStoreState>),
    <PkgRepoList />,
  );
  expect(wrapper.find(Alert)).toIncludeText("boom!");
});

it("shows an error deleting a repo", () => {
  const wrapper = mountWrapper(
    getStore({
      repos: { errors: { delete: new Error("boom!") } } as IPackageRepositoryState,
    } as Partial<IStoreState>),
    <PkgRepoList />,
  );
  expect(wrapper.find(Alert)).toIncludeText("boom!");
});

// TODO(andresmgot): Re-enable when the repo list is refactored
describe("global and namespaced repositories", () => {
  const globalRepo = {
    name: "bitnami",
    packageRepoRef: { context: { cluster: "default-cluster", namespace: helmGlobalNamespace } },
  } as PackageRepositorySummary;

  const namespacedRepo = {
    name: "my-repo",
    packageRepoRef: { context: { cluster: "default-cluster", namespace: namespace } },
    description: "my description 1 2 3 4",
  } as PackageRepositorySummary;

  it("shows a message if no global or namespaced repos exist", () => {
    const wrapper = mountWrapper(defaultStore, <PkgRepoList />);
    expect(
      wrapper
        .find("p")
        .filterWhere(p => p.text().includes("There are no global Package Repositories")),
    ).toExist();
    expect(
      wrapper
        .find("p")
        .filterWhere(p => p.text().includes("There are no namespaced Package Repositories")),
    ).toExist();
  });

  it("shows the global repositories with the buttons deactivated if the current user is not allowed to modify them", () => {
    Kube.canI = jest.fn().mockReturnValue({
      then: jest.fn((f: any) => f(false)),
      catch: jest.fn(f => f(false)),
    });
    const wrapper = mountWrapper(
      getStore({
        clusters: {
          ...initialState.clusters,
          clusters: {
            [currentCluster]: {
              ...initialState.clusters.clusters[initialState.clusters.currentCluster],
              currentNamespace: "other",
            },
          },
        },
        repos: {
          reposSummaries: [globalRepo],
        } as IPackageRepositoryState,
      } as Partial<IStoreState>),
      <PkgRepoList />,
    );
    expect(wrapper.find(Table)).toHaveLength(1);
    // The control buttons should be deactivated
    expect(wrapper.find(PkgRepoDisabledControl)).toExist();
    expect(wrapper.find(PkgRepoControl)).not.toExist();
    // The content related to namespaced repositories should exist
    expect(
      wrapper.find("h3").filterWhere(h => h.text().includes("Namespaced Repositories")),
    ).toExist();
  });

  it("shows the global repositories with the buttons enabled", () => {
    const wrapper = mountWrapper(
      getStore({
        clusters: {
          ...initialState.clusters,
          clusters: {
            [currentCluster]: {
              ...initialState.clusters.clusters[currentCluster],
              currentNamespace: helmGlobalNamespace,
            },
          },
        },
        repos: {
          reposSummaries: [globalRepo],
        } as IPackageRepositoryState,
      } as Partial<IStoreState>),
      <PkgRepoList />,
    );

    // A link to manage the repos should not exist since we are already there
    expect(wrapper.find("p").find(Link)).not.toExist();
    expect(wrapper.find(Table)).toHaveLength(1);
    // The control buttons should be enabled
    expect(wrapper.find(PkgRepoDisabledControl)).not.toExist();
    expect(wrapper.find(PkgRepoControl)).toExist();
    // The content related to namespaced repositories should be hidden
    expect(
      wrapper.find("h3").filterWhere(h => h.text().includes("Namespace Repositories")),
    ).not.toExist();
    // no tooltip for the global repo as it does not have a description.
    expect(wrapper.find(Tooltip)).not.toExist();
  });

  it("shows global and namespaced repositories", () => {
    const wrapper = mountWrapper(
      getStore({
        clusters: {
          ...initialState.clusters,
          clusters: {
            [currentCluster]: {
              ...initialState.clusters.clusters[currentCluster],
              currentNamespace: namespacedRepo.packageRepoRef?.context?.namespace,
            },
          },
        },
        repos: {
          reposSummaries: [globalRepo, namespacedRepo],
        } as IPackageRepositoryState,
      } as Partial<IStoreState>),
      <PkgRepoList />,
    );
    // A table per repository type
    expect(wrapper.find(TableRow)).toHaveLength(2);
  });

  it("shows a link to the repo catalog", () => {
    const wrapper = mountWrapper(
      getStore({
        repos: {
          reposSummaries: [namespacedRepo],
        } as IPackageRepositoryState,
      } as Partial<IStoreState>),
      <PkgRepoList />,
    );
    expect(wrapper.find(Table).find(Link).prop("to")).toEqual(
      `/c/${currentCluster}/ns/${namespacedRepo.packageRepoRef?.context?.namespace}/catalog?Repository=my-repo`,
    );
  });

  it("shows a tooltip for the repo", () => {
    const wrapper = mountWrapper(
      getStore({
        repos: {
          reposSummaries: [namespacedRepo],
        } as IPackageRepositoryState,
      } as Partial<IStoreState>),
      <PkgRepoList />,
    );
    act(() => {
      wrapper.find("input[type='checkbox']").simulate("change");
    });
    const tooltipText = wrapper.find(Tooltip).html();
    expect(tooltipText).toContain("my description 1 2 3 4");
  });

  it("use the correct namespace in the link when listing in all namespaces", () => {
    const wrapper = mountWrapper(
      getStore({
        repos: {
          reposSummaries: [namespacedRepo],
        } as IPackageRepositoryState,
      } as Partial<IStoreState>),
      <PkgRepoList />,
    );
    act(() => {
      wrapper.find("input[type='checkbox']").simulate("change");
    });
    expect(wrapper.find(Table).find(Link).prop("to")).toEqual(
      `/c/${currentCluster}/ns/${namespacedRepo.packageRepoRef?.context?.namespace}/catalog?Repository=my-repo`,
    );
  });
});
