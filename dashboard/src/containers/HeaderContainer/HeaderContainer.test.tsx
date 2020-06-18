import { shallow } from "enzyme";
import * as React from "react";
import { IAuthState } from "reducers/auth";
import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";
import Header from "./HeaderContainer";

const mockStore = configureMockStore([thunk]);

const emptyLocation = {
  hash: "",
  pathname: "",
  search: "",
};

const defaultAuthState: IAuthState = {
  sessionExpired: false,
  authenticated: true,
  oidcAuthenticated: true,
  authenticating: false,
  defaultNamespace: "",
};

const defaultState = {
  auth: defaultAuthState,
  router: { location: emptyLocation },
  config: {
    featureFlags: { operators: true },
  },
  clusters: {
    currentCluster: "default",
    clusters: {
      default: {
        currentNamespace: "default",
        namespaces: [],
      },
    },
  },
};

describe("HeaderContainer props", () => {
  it("maps authentication redux states to props", () => {
    const store = mockStore(defaultState);
    const wrapper = shallow(<Header store={store} />);
    const form = wrapper.find("Header");
    expect(form).toHaveProp({
      authenticated: true,
    });
  });

  it("maps featureFlags configuration to props", () => {
    const store = mockStore({
      ...defaultState,
      config: {
        featureFlags: { operators: true },
      },
    });

    const wrapper = shallow(<Header store={store} />);

    const form = wrapper.find("Header");
    expect(form).toHaveProp({
      featureFlags: { operators: true },
    });
  });
});
