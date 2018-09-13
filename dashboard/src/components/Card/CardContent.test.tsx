import { shallow } from "enzyme";
import * as React from "react";

import CardContent from "./CardContent";

it("should render the className", () => {
  const wrapper = shallow(<CardContent className="foo" />);
  expect(wrapper.find(".foo").exists()).toBe(true);
});

it("should render the children elements", () => {
  const children = <div>foo</div>;
  const wrapper = shallow(<CardContent children={children} />);
  expect(wrapper.text()).toContain("foo");
  expect(wrapper).toMatchSnapshot();
});
