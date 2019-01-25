import { shallow } from "enzyme";
import context from "jest-plugin-context";
import * as React from "react";

import { Link } from "react-router-dom";
import { hapi } from "shared/hapi/release";
import ChartInfo from "./ChartInfo";

const defaultProps = {
  app: {
    chart: {
      metadata: {
        appVersion: "0.0.1",
        description: "test chart",
        icon: "icon.png",
        version: "1.0.0",
      },
    },
    name: "foo",
  } as hapi.release.Release,
  updateCheck: { checked: false },
};

it("renders a app item", () => {
  const wrapper = shallow(<ChartInfo {...defaultProps} />);
  expect(wrapper.find(".ChartInfo").exists()).toBe(true);
  expect(wrapper).toMatchSnapshot();
});

context("when information about updates is available", () => {
  it("renders an up to date message if there are no updates", () => {
    const wrapper = shallow(
      <ChartInfo
        {...defaultProps}
        updateInfo={{ latestVersion: "", repository: { name: "", url: "" } }}
      />,
    );
    expect(wrapper.html()).toContain("Up to date");
  });
  it("renders an new version found message if the latest version is newer", () => {
    const wrapper = shallow(
      <ChartInfo
        {...defaultProps}
        updateInfo={{ latestVersion: "1.0.0", repository: { name: "", url: "" } }}
      />,
    );
    expect(
      wrapper
        .find(Link)
        .children()
        .text(),
    ).toContain("1.0.0 available");
  });
});
