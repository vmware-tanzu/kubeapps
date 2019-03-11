import { shallow } from "enzyme";
import * as React from "react";
import { Link } from "react-router-dom";

import ChartIcon from "../ChartIcon";
import ChartHeader from "./ChartHeader";

const testProps: any = {
  description: "A Test Chart",
  id: "testrepo/test",
  repo: "testrepo",
  version: {
    attributes: {
      app_version: "1.2.3",
    },
  },
};

it("renders a header for the chart", () => {
  const wrapper = shallow(<ChartHeader {...testProps} />);
  expect(wrapper.text()).toContain("testrepo/test");
  expect(wrapper.text()).toContain("A Test Chart");
  const repoLink = wrapper.find(Link);
  expect(repoLink.exists()).toBe(true);
  expect(repoLink.props()).toMatchObject({ to: "/catalog/testrepo", children: "testrepo" });
  expect(wrapper.find(ChartIcon).exists()).toBe(true);
  expect(wrapper).toMatchSnapshot();
});

it("displays the appVersion", () => {
  const wrapper = shallow(<ChartHeader {...testProps} />);
  expect(wrapper.text()).toContain("1.2.3");
});

it("uses the icon", () => {
  const wrapper = shallow(<ChartHeader {...testProps} icon="test.jpg" />);
  const icon = wrapper.find(ChartIcon);
  expect(icon.exists()).toBe(true);
  expect(icon.props()).toMatchObject({ icon: "test.jpg" });
});
