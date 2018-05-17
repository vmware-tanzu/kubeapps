import { shallow } from "enzyme";
import * as React from "react";

import { AlertTriangle, X } from "react-feather";
import ErrorAlertHeader from "./ErrorAlertHeader";

it("renders the heading passed to it", () => {
  const wrapper = shallow(<ErrorAlertHeader>test</ErrorAlertHeader>);
  expect(wrapper.text()).toContain("test");
  expect(wrapper).toMatchSnapshot();
});

it("renders the AlertTriangle icon by default", () => {
  const wrapper = shallow(<ErrorAlertHeader>test</ErrorAlertHeader>);
  expect(wrapper.find(".error__icon").contains(<AlertTriangle />)).toBe(true);
});

it("renders the icon passed in by prop", () => {
  const wrapper = shallow(<ErrorAlertHeader icon={X}>test</ErrorAlertHeader>);
  expect(wrapper.find(".error__icon").contains(<AlertTriangle />)).toBe(false);
  expect(wrapper.find(".error__icon").contains(<X />)).toBe(true);
});
