import { shallow } from "enzyme";
import * as React from "react";

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
  namespace: defaultNamespace,
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

  it("fetches repos when updating after a fetch error is cleared", () => {
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
});
