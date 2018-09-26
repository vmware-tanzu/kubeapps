import { shallow } from "enzyme";
import * as React from "react";

import NotFound from "./NotFound";

it("should render the 404 page", () => {
  const wrapper = shallow(<NotFound />);
  expect(wrapper.text()).toContain("The page you are looking for can't be found");
  expect(wrapper).toMatchSnapshot();
});
