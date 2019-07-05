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
  expect(button.text()).toBe("Deploy");
});

it("renders a redirect with the correct URL when the button is clicked", () => {
  const testCases = [
    { namespace: "test", version: "1.2.3", url: "/apps/ns/test/new/testrepo/test/versions/1.2.3" },
    {
      namespace: "foo",
      version: "alpha-0",
      url: "/apps/ns/foo/new/testrepo/test/versions/alpha-0",
    },
  ];

  testCases.forEach(t => {
    const chartVersion = Object.assign({}, testChartVersion);
    chartVersion.attributes.version = t.version;
    const wrapper = shallow(<ChartDeployButton version={chartVersion} namespace={t.namespace} />);
    const button = wrapper.find("button");
    expect(button.exists()).toBe(true);
    expect(wrapper.find(Redirect).exists()).toBe(false);

    button.simulate("click");
    const redirect = wrapper.find(Redirect);
    expect(redirect.exists()).toBe(true);
    expect(redirect.props()).toMatchObject({
      push: true,
      to: t.url,
    } as any);
  });
});
