import { shallow } from "enzyme";
import * as React from "react";
import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";

import RepoListContainer from ".";
import { definedNamespaces } from "../../shared/Namespace";

const mockStore = configureMockStore([thunk]);
const currentNamespace = "current-namespace";
const kubeappsNamespace = "kubeapps-namespace";

const defaultState = {
  config: {
    featureFlags: { reposPerNamespace: false },
    namespace: kubeappsNamespace,
  },
  namespace: { current: currentNamespace },
  repos: {},
};

describe("RepoListContainer props", () => {
  it("uses kubeapps namespace when reposPerNamespace false", () => {
    const store = mockStore(defaultState);
    const wrapper = shallow(<RepoListContainer store={store} />);

    const component = wrapper.find("AppRepoList");

    expect(component).toHaveProp({
      namespace: kubeappsNamespace,
    });
  });

  it("uses the current namespace when reposPerNamespace is true", () => {
    const store = mockStore({
      ...defaultState,
      config: {
        featureFlags: { reposPerNamespace: true },
      },
    });
    const wrapper = shallow(<RepoListContainer store={store} />);

    const component = wrapper.find("AppRepoList");

    expect(component).toHaveProp({
      namespace: currentNamespace,
    });
  });

  it("uses kubeapps namespace when reposPerNamespace is true and _all namespaces selected", () => {
    const store = mockStore({
      ...defaultState,
      config: {
        featureFlags: { reposPerNamespace: true },
        namespace: kubeappsNamespace,
      },
      namespace: { current: definedNamespaces.all },
    });
    const wrapper = shallow(<RepoListContainer store={store} />);

    const component = wrapper.find("AppRepoList");

    expect(component).toHaveProp({
      namespace: kubeappsNamespace,
    });
  });
});
