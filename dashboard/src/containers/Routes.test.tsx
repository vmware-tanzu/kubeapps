import { mount } from "enzyme";
import * as React from "react";
import { Redirect, StaticRouter } from "react-router";
import NotFound from "../components/NotFound";
import Routes from "./Routes";

it("invalid path should show a 404 error", () => {
  const wrapper = mount(
    <StaticRouter location="/random" context={{}}>
      <Routes />
    </StaticRouter>,
  );
  expect(wrapper.find(NotFound)).toExist();
  expect(wrapper.text()).toContain("The page you are looking for can't be found.");
});

it("should render a redirect to the default page", () => {
  const wrapper = mount(
    <StaticRouter location="/" context={{}}>
      <Routes />
    </StaticRouter>,
  );
  expect(wrapper.find(NotFound)).not.toExist();
  expect(wrapper.find(Redirect).props().to).toEqual("/apps/ns/default");
});
