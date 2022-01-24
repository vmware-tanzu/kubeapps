// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { mount } from "enzyme";
import Tabs from "./Tabs";

it("renders several tabs", () => {
  const wrapper = mount(<Tabs id="tabs" columns={["foo", "bar"]} data={["FOO", "BAR"]} />);
  expect(wrapper.find("button")).toHaveLength(2);
  expect(wrapper.find("section")).toHaveLength(2);
  expect(wrapper).toMatchSnapshot();
});

it("changes content between tabs", () => {
  const wrapper = mount(<Tabs id="tabs" columns={["foo", "bar"]} data={["FOO", "BAR"]} />);
  expect(wrapper.find(".active").text()).toEqual("foo");
  expect(
    wrapper
      .find("section")
      .findWhere(s => s.prop("aria-hidden") === "false")
      .text(),
  ).toEqual("FOO");

  const otherTab = wrapper.find("#tabs-tab1");
  otherTab.simulate("click");
  wrapper.update();

  const newSelectedTab = wrapper.find("#tabs-tab1");
  expect(newSelectedTab).toHaveClassName("active");
  expect(newSelectedTab.text()).toEqual("bar");
  expect(
    wrapper
      .find("section")
      .findWhere(s => s.prop("aria-hidden") === "false")
      .text(),
  ).toEqual("BAR");
});
