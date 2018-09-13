import { shallow } from "enzyme";
import * as React from "react";

import CardGrid from "./CardGrid";

it("should render the className", () => {
  const wrapper = shallow(<CardGrid className="foo" />);
  expect(wrapper.find(".foo").exists()).toBe(true);
});

it("should render the children elements", () => {
  const children = <div>foo</div>;
  const wrapper = shallow(<CardGrid children={children} />);
  expect(wrapper.text()).toContain("foo");
  expect(wrapper).toMatchSnapshot();
});
