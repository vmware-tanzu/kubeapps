// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { mount } from "enzyme";
import OperatorHeader from "./OperatorHeader";

const defaultProps = {
  title: "foo by Kubeapps",
  icon: "/path/to/icon.png",
  version: "1.0.0",
};

it("fallbacks to the default icon if not set", () => {
  const wrapper = mount(<OperatorHeader {...defaultProps} icon={undefined} />);
  expect(
    wrapper
      .find("img")
      .filterWhere(i => i.prop("alt") === "icon")
      .prop("src"),
  ).toBe("placeholder.svg");
});

it("includes the id, provider and version", () => {
  const wrapper = mount(<OperatorHeader {...defaultProps} />);
  expect(wrapper).toIncludeText("foo by Kubeapps");
  expect(wrapper).toIncludeText("Operator Version: 1.0.0");
});

it("renders buttons", () => {
  const wrapper = mount(
    <OperatorHeader {...defaultProps} buttons={[<div key="foo" id="foo" />]} />,
  );
  expect(wrapper.find("#foo")).toExist();
});
