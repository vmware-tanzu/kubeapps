import { mount } from "enzyme";
import { createMemoryHistory } from "history";
import * as React from "react";
import { StaticRouter } from "react-router";
import { Redirect, RouteComponentProps } from "react-router-dom";

import NotFound from "../../components/NotFound";
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
      <Routes {...emptyRouteComponentProps} namespace={"default"} />
    </StaticRouter>,
  );
  expect(wrapper.find(NotFound)).toExist();
  expect(wrapper.text()).toContain("The page you are looking for can't be found.");
});

it("should render a redirect to the default page", () => {
  const wrapper = mount(
    <StaticRouter location="/" context={{}}>
      <Routes {...emptyRouteComponentProps} namespace={"default"} />
    </StaticRouter>,
  );
  expect(wrapper.find(NotFound)).not.toExist();
  expect(wrapper.find(Redirect).props().to).toEqual("/apps/ns/default");
});
