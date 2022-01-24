// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { shallow } from "enzyme";
import NotFound from "./NotFound";

it("should render the 404 page", () => {
  const wrapper = shallow(<NotFound />);
  expect(wrapper.text()).toContain("The page you are looking for can't be found");
  expect(wrapper).toMatchSnapshot();
});
