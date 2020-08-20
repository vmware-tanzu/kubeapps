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
    namespace: kubeappsNamespace,
    featureFlags: { ui: "clarity" },
  },
  clusters: {
    currentCluster: "default",
    clusters: {
      default: { currentNamespace },
    },
  },
  repos: {},
  displayReposPerNamespaceMsg: false,
};

describe("RepoListContainer props", () => {
  it("uses the current namespace", () => {
    const store = mockStore({
      ...defaultState,
    });
    const wrapper = shallow(<RepoListContainer store={store} />);

    const component = wrapper.find("AppRepoListSelector");

    expect(component).toHaveProp({
      namespace: currentNamespace,
      displayReposPerNamespaceMsg: true,
    });
  });

  it("passes _all through as a normal namespace to be handled by the component", () => {
    const store = mockStore({
      ...defaultState,
      clusters: {
        ...defaultState.clusters,
        clusters: {
          default: { currentNamespace: definedNamespaces.all },
        },
      },
    });
    const wrapper = shallow(<RepoListContainer store={store} />);

    const component = wrapper.find("AppRepoListSelector");

    expect(component).toHaveProp({
      namespace: definedNamespaces.all,
      displayReposPerNamespaceMsg: false,
    });
  });
});
