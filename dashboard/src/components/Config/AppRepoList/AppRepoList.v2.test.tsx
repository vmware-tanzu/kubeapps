import actions from "actions";
import Alert from "components/js/Alert";
import PageHeader from "components/PageHeader/PageHeader.v2";
import * as React from "react";
import * as ReactRedux from "react-redux";
import { defaultStore, getStore, mountWrapper } from "shared/specs/mountWrapper";
import { AppRepoAddButton } from "./AppRepoButton.v2";
import AppRepoList from "./AppRepoList.v2";
import { AppRepoRefreshAllButton } from "./AppRepoRefreshAllButton.v2";

const defaultNamespace = "default-namespace";

const defaultProps = {
  namespace: defaultNamespace,
  cluster: "default",
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
  expect(actions.repos.fetchRepos).toHaveBeenCalledWith(defaultProps.namespace);
  expect(actions.repos.fetchImagePullSecrets).toHaveBeenCalledWith(defaultProps.namespace);
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

it("shows an error updating a repo", () => {
  const wrapper = mountWrapper(
    getStore({ repos: { errors: { update: new Error("boom!") } } }),
    <AppRepoList {...defaultProps} />,
  );
  expect(wrapper.find(Alert)).toIncludeText("boom!");
});

// TODO(andresmgot): Add test for the tables once global/namespaced repositories
// are implemented.
