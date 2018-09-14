import { shallow } from "enzyme";
import * as React from "react";

import CardGrid from "./CardGrid";

it("should render the className", () => {
  const wrapper = shallow(<CardGrid className="foo" />);
  expect(wrapper.find(".foo")).toExist();
});

it("should render the children elements", () => {
  const wrapper = shallow(
    <CardGrid>
      <div>foo</div>
    </CardGrid>,
  );
  expect(wrapper.text()).toContain("foo");
  expect(wrapper).toMatchSnapshot();
});
