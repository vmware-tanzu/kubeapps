import { shallow } from "enzyme";
import * as React from "react";

import CardFooter from "./CardFooter";

it("should render the className", () => {
  const wrapper = shallow(<CardFooter className="foo" />);
  expect(wrapper.find(".foo")).toExist();
});

it("should render the children elements", () => {
  const wrapper = shallow(
    <CardFooter>
      <div>foo</div>
    </CardFooter>,
  );
  expect(wrapper.text()).toContain("foo");
  expect(wrapper).toMatchSnapshot();
});
