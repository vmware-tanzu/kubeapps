import actions from "actions";
import Alert from "components/js/Alert";
import Table from "components/js/Table";
import PageHeader from "components/PageHeader/PageHeader.v2";
import * as React from "react";
import * as ReactRedux from "react-redux";
import { Link } from "react-router-dom";
import { defaultStore, getStore, mountWrapper } from "shared/specs/mountWrapper";
import { app } from "shared/url";
import { AppRepoAddButton } from "./AppRepoButton.v2";
import { AppRepoControl } from "./AppRepoControl.v2";
import { AppRepoDisabledControl } from "./AppRepoDisabledControl.v2";
import AppRepoList from "./AppRepoList.v2";
import { AppRepoRefreshAllButton } from "./AppRepoRefreshAllButton.v2";

const defaultNamespace = "default-namespace";

const defaultProps = {
  namespace: defaultNamespace,
  cluster: "default",
  kubeappsCluster: "default",
  kubeappsNamespace: "kubeapps",
};

let spyOnUseDispatch: jest.SpyInstance;
const kubeaActions = { ...actions.kube };
beforeEach(() => {
  actions.repos = {
    ...actions.repos,
    fetchRepos: jest.fn(),
    fetchImagePullSecrets: jest.fn(),
  };
  const mockDispatch = jest.fn();
  spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
});

afterEach(() => {
  actions.kube = { ...kubeaActions };
  spyOnUseDispatch.mockRestore();
});

it("fetches repos and imagePullSecrets", () => {
  mountWrapper(defaultStore, <AppRepoList {...defaultProps} />);
  expect(actions.repos.fetchRepos).toHaveBeenCalledWith(
    defaultProps.namespace,
    defaultProps.kubeappsNamespace,
  );
  expect(actions.repos.fetchImagePullSecrets).toHaveBeenCalledWith(defaultProps.namespace);
});

it("fetches repos only from the kubeappsNamespace", () => {
  mountWrapper(
    defaultStore,
    <AppRepoList {...defaultProps} namespace={defaultProps.kubeappsNamespace} />,
  );
  expect(actions.repos.fetchRepos).toHaveBeenCalledWith(defaultProps.kubeappsNamespace);
});

// TODO: Remove this test when app repos are supported in different clusters
it("shows a warning if the cluster is not the default one", () => {
  const wrapper = mountWrapper(defaultStore, <AppRepoList {...defaultProps} cluster="other" />);
  expect(wrapper.find(Alert)).toIncludeText(
    "App Repositories are available on the default cluster only",
  );
});

it("renders the button to add a repo and refresh all", () => {
  const wrapper = mountWrapper(defaultStore, <AppRepoList {...defaultProps} />);
  expect(wrapper.find(PageHeader).find(AppRepoAddButton)).toExist();
  expect(wrapper.find(PageHeader).find(AppRepoRefreshAllButton)).toExist();
});

it("shows an error fetching a repo", () => {
  const wrapper = mountWrapper(
    getStore({ repos: { errors: { fetch: new Error("boom!") } } }),
    <AppRepoList {...defaultProps} />,
  );
  expect(wrapper.find(Alert)).toIncludeText("boom!");
});

it("shows an error deleting a repo", () => {
  const wrapper = mountWrapper(
    getStore({ repos: { errors: { delete: new Error("boom!") } } }),
    <AppRepoList {...defaultProps} />,
  );
  expect(wrapper.find(Alert)).toIncludeText("boom!");
});

describe("global and namespaced repositories", () => {
  const globalRepo = {
    metadata: {
      name: "bitnami",
      namespace: defaultProps.kubeappsNamespace,
    },
    spec: {},
  };

  const namespacedRepo = {
    metadata: {
      name: "my-repo",
      namespace: defaultProps.namespace,
    },
    spec: {},
  };

  it("shows a message if no global or namespaced repos exist", () => {
    const wrapper = mountWrapper(defaultStore, <AppRepoList {...defaultProps} />);
    expect(
      wrapper.find("p").filterWhere(p => p.text().includes("No global repositories found")),
    ).toExist();
    expect(
      wrapper
        .find("p")
        .filterWhere(p => p.text().includes("The current namespace doesn't have any repositories")),
    ).toExist();
  });

  it("shows the global repositories with the buttons disabled if the current namespace is other", () => {
    const wrapper = mountWrapper(
      getStore({
        repos: {
          repos: [globalRepo],
        },
      }),
      <AppRepoList {...defaultProps} namespace="other" />,
    );

    // A link to manage the repos should exist
    expect(wrapper.find(Link).prop("to")).toBe(
      app.config.apprepositories("default", defaultProps.kubeappsNamespace),
    );
    expect(wrapper.find(Table)).toHaveLength(1);
    // The control buttons should be disabled
    expect(wrapper.find(AppRepoDisabledControl)).toExist();
    expect(wrapper.find(AppRepoControl)).not.toExist();
    // The content related to namespaced repositories should exist
    expect(
      wrapper.find("h3").filterWhere(h => h.text().includes("Namespace Repositories")),
    ).toExist();
  });

  it("shows the global repositories with the buttons enabled", () => {
    const wrapper = mountWrapper(
      getStore({
        repos: {
          repos: [globalRepo],
        },
      }),
      <AppRepoList {...defaultProps} namespace={defaultProps.kubeappsNamespace} />,
    );

    // A link to manage the repos should not exist since we are already there
    expect(wrapper.find(Link)).not.toExist();
    expect(wrapper.find(Table)).toHaveLength(1);
    // The control buttons should be enabled
    expect(wrapper.find(AppRepoDisabledControl)).not.toExist();
    expect(wrapper.find(AppRepoControl)).toExist();
    // The content related to namespaced repositories should be hidden
    expect(
      wrapper.find("h3").filterWhere(h => h.text().includes("Namespace Repositories")),
    ).not.toExist();
  });

  it("shows global and namespaced repositories", () => {
    const wrapper = mountWrapper(
      getStore({
        repos: {
          repos: [globalRepo, namespacedRepo],
        },
      }),
      <AppRepoList {...defaultProps} namespace={namespacedRepo.metadata.namespace} />,
    );

    // A table per repository type
    expect(wrapper.find(Table)).toHaveLength(2);
    // The control buttons should be enabled for the namespaced repository and disabled
    // for the global one
    expect(
      wrapper
        .find(Table)
        .at(0)
        .find(AppRepoDisabledControl),
    ).toExist();
    expect(
      wrapper
        .find(Table)
        .at(1)
        .find(AppRepoControl),
    ).toExist();
  });
});
