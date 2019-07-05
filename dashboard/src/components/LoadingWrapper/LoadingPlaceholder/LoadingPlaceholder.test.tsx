import { shallow } from "enzyme";
import * as React from "react";
import LoadingPlaceholder from "./LoadingPlaceholder";

it("matches the snapshot", () => {
  const wrapper = shallow(<LoadingPlaceholder />);
  expect(wrapper).toMatchSnapshot();
});
