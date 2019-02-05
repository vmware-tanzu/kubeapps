import { shallow } from "enzyme";
import context from "jest-plugin-context";
import * as React from "react";

import { Link } from "react-router-dom";
import { hapi } from "shared/hapi/release";
import { IRelease } from "shared/types";
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
};

it("renders a app item", () => {
  const wrapper = shallow(<ChartInfo {...defaultProps} />);
  expect(wrapper.find(".ChartInfo").exists()).toBe(true);
  expect(wrapper).toMatchSnapshot();
});

context("when information about updates is available", () => {
  it("renders an up to date message if there are no updates", () => {
    const appWithoutUpdates = { ...defaultProps.app, updateInfo: {} } as IRelease;
    const wrapper = shallow(<ChartInfo {...defaultProps} app={appWithoutUpdates} />);
    expect(wrapper.html()).toContain("Up to date");
  });
  it("renders an new version found message if the latest version is newer", () => {
    const appWithUpdates = {
      ...defaultProps.app,
      updateInfo: { latestVersion: "1.0.0" },
    } as IRelease;
    const wrapper = shallow(<ChartInfo {...defaultProps} app={appWithUpdates} />);
    expect(
      wrapper
        .find(Link)
        .children()
        .text(),
    ).toContain("1.0.0 available");
  });
});
