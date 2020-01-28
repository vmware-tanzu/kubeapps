import { shallow } from "enzyme";
import * as React from "react";
import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";
import AppUpgrade from "../../components/AppUpgrade";
import Upgrade from "./AppUpgradeContainer";

const mockStore = configureMockStore([thunk]);

const emptyLocation = {
  hash: "",
  pathname: "",
  search: "",
};

const makeStore = (apps: any, repos: any) => {
  return mockStore({
    apps,
    repos,
    router: { location: emptyLocation },
    charts: { selected: {} },
    config: {},
  });
};

describe("LoginFormContainer props", () => {
  it("repoName is empty if no apps nor repos are available", () => {
    const store = makeStore({}, { errors: {}, repo: {} });
    const match = { params: { namespace: "default", releaseName: "foo" } };
    const wrapper = shallow(<Upgrade store={store} match={match} />);
    expect(wrapper.find(AppUpgrade).prop("repoName")).toBe(undefined);
  });

  it("repoName is set using the selected repo", () => {
    const store = makeStore({}, { errors: {}, repo: { metadata: { name: "stable" } } });
    const match = { params: { namespace: "default", releaseName: "foo" } };
    const wrapper = shallow(<Upgrade store={store} match={match} />);
    expect(wrapper.find(AppUpgrade).prop("repoName")).toBe("stable");
  });

  it("repoName is set using the updateInfo", () => {
    const store = makeStore(
      { selected: { updateInfo: { repository: { name: "bitnami" } } } },
      { errors: {}, repo: {} },
    );
    const match = { params: { namespace: "default", releaseName: "foo" } };
    const wrapper = shallow(<Upgrade store={store} match={match} />);
    expect(wrapper.find(AppUpgrade).prop("repoName")).toBe("bitnami");
  });
});
