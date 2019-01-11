import { shallow } from "enzyme";
import * as React from "react";

import ServiceItem from "./ServiceItem";
import ServiceTable from "./ServicesTable";

it("renders a message if there are no services or ingresses", () => {
  const wrapper = shallow(<ServiceTable serviceRefs={[]} />);
  expect(wrapper.find(ServiceItem)).not.toExist();
  expect(wrapper.text()).toContain("The current application does not contain any Service objects");
});
