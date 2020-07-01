import { mount } from "enzyme";
import { createMemoryHistory } from "history";
import * as React from "react";
import { StaticRouter } from "react-router";
import { Redirect, RouteComponentProps } from "react-router-dom";
import NotFound from "../../components/NotFound";
import { app } from "../../shared/url";
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
      <Routes
        {...emptyRouteComponentProps}
        cluster={"default"}
        namespace={"default"}
        authenticated={true}
      />
    </StaticRouter>,
  );
  expect(wrapper.find(NotFound)).toExist();
  expect(wrapper.text()).toContain("The page you are looking for can't be found.");
});

it("should render a redirect to the default cluster and namespace", () => {
  const wrapper = mount(
    <StaticRouter location="/" context={{}}>
      <Routes
        {...emptyRouteComponentProps}
        cluster={"default"}
        namespace={"default"}
        authenticated={true}
      />
    </StaticRouter>,
  );
  expect(wrapper.find(NotFound)).not.toExist();
  expect(wrapper.find(Redirect).prop("to")).toEqual(app.apps.list("default", "default"));
});

it("should render a redirect to the login page", () => {
  const wrapper = mount(
    <StaticRouter location="/" context={{}}>
      <Routes
        {...emptyRouteComponentProps}
        cluster={"default"}
        namespace={""}
        authenticated={true}
      />
    </StaticRouter>,
  );
  expect(wrapper.find(NotFound)).not.toExist();
  expect(wrapper.find(Redirect).prop("to")).toEqual("/login");
});

it("should render a redirect to the login page (when not authenticated)", () => {
  const wrapper = mount(
    <StaticRouter location="/" context={{}}>
      <Routes
        {...emptyRouteComponentProps}
        cluster={"default"}
        namespace={"default"}
        authenticated={false}
      />
    </StaticRouter>,
  );
  expect(wrapper.find(NotFound)).not.toExist();
  expect(wrapper.find(Redirect).prop("to")).toEqual("/login");
});
