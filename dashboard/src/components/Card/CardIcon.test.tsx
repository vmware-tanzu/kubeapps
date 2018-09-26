import { shallow } from "enzyme";
import * as React from "react";

import CardIcon from "./CardIcon";

it("should return an empty object if the icon is not set", () => {
  const wrapper = shallow(<CardIcon />);
  expect(wrapper.html()).toBe(null);
});

it("should render the icon if set", () => {
  const wrapper = shallow(<CardIcon icon="foo" />);
  expect(wrapper.html()).not.toBe(null);
  expect(wrapper.find("img").props().src).toBe("foo");
  expect(wrapper).toMatchSnapshot();
});
