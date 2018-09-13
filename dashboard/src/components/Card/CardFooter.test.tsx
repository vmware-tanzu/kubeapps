import { shallow } from "enzyme";
import * as React from "react";

import CardFooter from "./CardFooter";

it("should render the className", () => {
  const wrapper = shallow(<CardFooter className="foo" />);
  expect(wrapper.find(".foo").exists()).toBe(true);
});

it("should render the children elements", () => {
  const children = <div>foo</div>;
  const wrapper = shallow(<CardFooter children={children} />);
  expect(wrapper.text()).toContain("foo");
  expect(wrapper).toMatchSnapshot();
});
