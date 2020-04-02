import { mount, shallow } from "enzyme";
import { createMemoryHistory } from "history";
import * as React from "react";
import { StaticRouter } from "react-router";
import { Redirect, RouteComponentProps } from "react-router-dom";

import NotFound from "../../components/NotFound";
import RepoListContainer from "../../containers/RepoListContainer";
import Routes from "./Routes";

const emptyRouteComponentProps: RouteComponentProps<{}> = {
  history: createMemoryHistory(),
  location: {
    hash: "",
    pathname: "",
    search: "",
    state: "",
  },
  match: {
    isExact: false,
    params: {},
    path: "",
    url: "",
  },
};

it("invalid path should show a 404 error", () => {
  const wrapper = mount(
    <StaticRouter location="/random" context={{}}>
      <Routes {...emptyRouteComponentProps} namespace={"default"} authenticated={true} />
    </StaticRouter>,
  );
  expect(wrapper.find(NotFound)).toExist();
  expect(wrapper.text()).toContain("The page you are looking for can't be found.");
});

it("should render a redirect to the default namespace", () => {
  const wrapper = mount(
    <StaticRouter location="/" context={{}}>
      <Routes {...emptyRouteComponentProps} namespace={"default"} authenticated={true} />
    </StaticRouter>,
  );
  expect(wrapper.find(NotFound)).not.toExist();
  expect(wrapper.find(Redirect).prop("to")).toEqual("/ns/default/apps");
});

it("should render a redirect to the login page", () => {
  const wrapper = mount(
    <StaticRouter location="/" context={{}}>
      <Routes {...emptyRouteComponentProps} namespace={""} authenticated={true} />
    </StaticRouter>,
  );
  expect(wrapper.find(NotFound)).not.toExist();
  expect(wrapper.find(Redirect).prop("to")).toEqual("/login");
});

it("should render a redirect to the login page (when not authenticated)", () => {
  const wrapper = mount(
    <StaticRouter location="/" context={{}}>
      <Routes {...emptyRouteComponentProps} namespace={"default"} authenticated={false} />
    </StaticRouter>,
  );
  expect(wrapper.find(NotFound)).not.toExist();
  expect(wrapper.find(Redirect).prop("to")).toEqual("/login");
});

describe("Routes depending on feature flags", () => {
  const namespace = "default";
  const perNamespacePath = "/config/ns/:namespace/repos";
  const nonNamespacedPath = "/config/repos";

  it("should use a non-namespaced route for app repos without feature flag", () => {
    const wrapper = shallow(
      <StaticRouter location="/config/repos" context={{}}>
        <Routes
          {...emptyRouteComponentProps}
          namespace={namespace}
          authenticated={true}
          featureFlags={{ reposPerNamespace: false, operators: false }}
        />
      </StaticRouter>,
    )
      .dive()
      .dive()
      .dive();

    const component = wrapper.find({ component: RepoListContainer });

    expect(component.length).toBe(1);
    expect(component.props().path).toEqual(nonNamespacedPath);
  });

  it("should use a namespaced route for app repos when feature flag set", () => {
    const wrapper = shallow(
      <StaticRouter location="/config/repos" context={{}}>
        <Routes
          {...emptyRouteComponentProps}
          namespace={namespace}
          authenticated={true}
          featureFlags={{ reposPerNamespace: true, operators: false }}
        />
      </StaticRouter>,
    )
      .dive()
      .dive()
      .dive();

    const component = wrapper.find({ component: RepoListContainer });

    expect(component.length).toBe(1);
    expect(component.props().path).toEqual(perNamespacePath);
  });
});
