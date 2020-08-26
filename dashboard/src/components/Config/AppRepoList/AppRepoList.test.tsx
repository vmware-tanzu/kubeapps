import { shallow } from "enzyme";
import * as React from "react";

import { MessageAlert } from "components/ErrorAlert";
import AppRepoList from "./AppRepoList";

const defaultNamespace = "default-namespace";

const defaultProps = {
  errors: {},
  repos: [],
  fetchRepos: jest.fn(),
  deleteRepo: jest.fn(),
  resyncRepo: jest.fn(),
  resyncAllRepos: jest.fn(),
  install: jest.fn(),
  update: jest.fn(),
  validate: jest.fn(),
  namespace: defaultNamespace,
  cluster: "default",
  kubeappsCluster: "default",
  kubeappsNamespace: "kubeapps",
  displayReposPerNamespaceMsg: false,
  validating: false,
  isFetching: false,
  repoSecrets: [],
  fetchImagePullSecrets: jest.fn(),
  imagePullSecrets: [],
  createDockerRegistrySecret: jest.fn(),
};

describe("AppRepoList", () => {
  it("fetches repos for a namespace when mounted", () => {
    const props = {
      ...defaultProps,
      fetchRepos: jest.fn(),
    };

    shallow(<AppRepoList {...props} />);

    expect(props.fetchRepos).toHaveBeenCalledWith(defaultNamespace);
  });

  it("refetches repos when updating after a fetch error is cleared", () => {
    const props = {
      ...defaultProps,
      errors: { fetch: new Error("Bang!") },
      fetchRepos: jest.fn(),
    };

    const wrapper = shallow(<AppRepoList {...props} />);
    wrapper.setProps({
      ...props,
      errors: {},
    });

    expect(props.fetchRepos).toHaveBeenCalledTimes(2);
    expect(props.fetchRepos).toHaveBeenLastCalledWith(defaultNamespace);
  });

  it("refetches repos when the namespace changes", () => {
    const props = {
      ...defaultProps,
      fetchRepos: jest.fn(),
    };
    const differentNamespace = "different-namespace";

    const wrapper = shallow(<AppRepoList {...props} />);
    wrapper.setProps({
      ...props,
      namespace: differentNamespace,
    });

    expect(props.fetchRepos).toHaveBeenCalledTimes(2);
    expect(props.fetchRepos).toHaveBeenLastCalledWith(differentNamespace);
  });

  it("does not refetch otherwise", () => {
    const props = {
      ...defaultProps,
      fetchRepos: jest.fn(),
    };

    const wrapper = shallow(<AppRepoList {...props} />);
    wrapper.setProps({
      ...props,
      errors: { fetch: new Error("Bang!") },
    });

    expect(props.fetchRepos).toHaveBeenCalledTimes(1);
  });

  it("displays LoadingWrapper when fetching", () => {
    const props = {
      ...defaultProps,
      isFetching: true,
    };

    const wrapper = shallow(<AppRepoList {...props} />);

    const loading = wrapper.find("LoadingWrapper");
    expect(loading.length).toBe(1);
    expect(loading).toHaveProp({
      loaded: false,
    });
  });

  it("displays the AppRepoAddButton when no fetching errors", () => {
    const wrapper = shallow(<AppRepoList {...defaultProps} />);

    const addButton = wrapper.find("AppRepoAddButton");
    expect(addButton.length).toBe(1);
  });

  it("does not display the AppRepoAddButton when there is a fetching error", () => {
    const props = {
      ...defaultProps,
      errors: { fetch: new Error("Bang!") },
    };
    const wrapper = shallow(<AppRepoList {...props} />);

    const addButton = wrapper.find("AppRepoAddButton");
    expect(addButton.length).toBe(0);
  });

  it("displays not-supported alert when an additional cluster is selected", () => {
    const props = {
      ...defaultProps,
      cluster: "other-cluster",
    };

    const wrapper = shallow(<AppRepoList {...props} />);

    const msgAlert = wrapper.find(MessageAlert);
    expect(msgAlert).toExist();
    expect(msgAlert.prop("header")).toEqual(
      "AppRepositories can be created on the default cluster only",
    );
    const addButton = wrapper.find("AppRepoAddButton");
    expect(addButton.length).toBe(0);
  });
});
