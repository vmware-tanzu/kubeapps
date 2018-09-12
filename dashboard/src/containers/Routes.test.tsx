import { mount } from "enzyme";
import * as React from "react";
import { MemoryRouter } from "react-router";
import NotFound from "../components/NotFound";
import Routes from "./Routes";

it("invalid path should redirect to 404", () => {
  const wrapper = mount(
    <MemoryRouter initialEntries={["/random"]}>
      <Routes namespace="default" />
    </MemoryRouter>,
  );
  expect(wrapper.find(NotFound).exists()).toBe(true);
  expect(wrapper.text()).toContain("The page you are looking for can't be found.");
});
