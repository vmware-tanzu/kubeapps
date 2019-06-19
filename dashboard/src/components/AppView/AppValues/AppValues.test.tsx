import { shallow } from "enzyme";
import * as React from "react";
import AppValues from "./";

it("match snapshot with values", () => {
  const wrapper = shallow(<AppValues values="foo: bar" />);
  expect(wrapper).toMatchSnapshot();
});

it("match snapshot without values", () => {
  const wrapper = shallow(<AppValues values="" />);
  expect(wrapper).toMatchSnapshot();
});
