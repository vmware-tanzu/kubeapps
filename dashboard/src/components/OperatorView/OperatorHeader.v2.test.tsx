import { mount, shallow } from "enzyme";
import * as React from "react";
import OperatorHeader from "./OperatorHeader.v2";

const defaultProps = {
  title: "foo by Kubeapps",
  icon: "/path/to/icon.png",
  version: "1.0.0",
};

it("fallbacks to the default icon if not set", () => {
  const wrapper = shallow(<OperatorHeader {...defaultProps} icon={undefined} />);
  expect(
    wrapper
      .find("img")
      .filterWhere(i => i.prop("alt") === "app-icon")
      .prop("src"),
  ).toBe("placeholder.png");
});

it("includes the id, provider and version", () => {
  const wrapper = mount(<OperatorHeader {...defaultProps} />);
  expect(wrapper).toIncludeText("foo by Kubeapps");
  expect(wrapper).toIncludeText("Operator Version: 1.0.0");
});

it("renders children component", () => {
  const wrapper = shallow(
    <OperatorHeader {...defaultProps}>
      <div id="foo" />
    </OperatorHeader>,
  );
  expect(wrapper.find("#foo")).toExist();
});
