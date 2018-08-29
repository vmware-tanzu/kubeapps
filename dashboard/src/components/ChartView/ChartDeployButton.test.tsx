import { shallow } from "enzyme";
import * as React from "react";
import { Redirect } from "react-router";

import { IChartVersion } from "../../shared/types";
import ChartDeployButton from "./ChartDeployButton";

const testChartVersion: IChartVersion = {
  attributes: {
    version: "1.2.3",
  },
  relationships: {
    chart: {
      data: {
        name: "test",
        repo: {
          name: "testrepo",
        },
      },
    },
  },
} as IChartVersion;

it("renders a button to deploy the chart version", () => {
  const wrapper = shallow(<ChartDeployButton version={testChartVersion} namespace="test" />);
  const button = wrapper.find("button");
  expect(button.exists()).toBe(true);
  expect(button.text()).toBe("Deploy using Helm");
});

it("renders a redirect when the button is clicked", () => {
  const wrapper = shallow(<ChartDeployButton version={testChartVersion} namespace="test" />);
  const button = wrapper.find("button");
  expect(button.exists()).toBe(true);
  expect(wrapper.find(Redirect).exists()).toBe(false);

  button.simulate("click");
  const redirect = wrapper.find(Redirect);
  expect(redirect.exists()).toBe(true);
  expect(redirect.props()).toMatchObject({
    push: true,
    to: "/apps/ns/test/new/testrepo/test/versions/1.2.3",
  });
});
