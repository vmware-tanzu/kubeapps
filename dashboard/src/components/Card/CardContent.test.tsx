import { shallow } from "enzyme";
import * as React from "react";

import CardContent from "./CardContent";

it("should render the className", () => {
  const wrapper = shallow(<CardContent className="foo" />);
  expect(wrapper.find(".foo")).toExist();
});

it("should render the children elements", () => {
  const wrapper = shallow(
    <CardContent>
      <div>foo</div>
    </CardContent>,
  );
  expect(wrapper.text()).toContain("foo");
  expect(wrapper).toMatchSnapshot();
});
