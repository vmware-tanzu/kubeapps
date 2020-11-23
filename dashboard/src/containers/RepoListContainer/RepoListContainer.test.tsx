import { shallow } from "enzyme";
import * as React from "react";
import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";
import RepoListContainer from ".";

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
};

describe("RepoListContainer props", () => {
  it("uses the current namespace", () => {
    const store = mockStore({
      ...defaultState,
    });
    const wrapper = shallow(<RepoListContainer store={store} />);

    const component = wrapper.find("AppRepoList");

    expect(component).toHaveProp({
      namespace: currentNamespace,
    });
  });
});
