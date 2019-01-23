import { mount, shallow } from "enzyme";
import context from "jest-plugin-context";
import * as React from "react";
import { Redirect } from "react-router";

import UpgradeButton from "./UpgradeButton";

it("renders a redirect when clicking upgrade", () => {
  const updateURL = "/apps/ns/default/upgrade/foo";
  const wrapper = shallow(<UpgradeButton upgradeURL={updateURL} />);
  const button = wrapper.find(".button").filterWhere(i => i.text() === "Upgrade");
  expect(button.exists()).toBe(true);
  expect(wrapper.find(Redirect).exists()).toBe(false);

  button.simulate("click");
  const redirect = wrapper.find(Redirect);
  expect(redirect.exists()).toBe(true);
  expect(redirect.props()).toMatchObject({
    push: true,
    to: updateURL,
  });
});

context("when a new version is available", () => {
  const update = { checked: true, repository: { name: "foo", url: "" }, latestVersion: "1.0.0" };
  it("should show a tooltip to notify the new version", () => {
    const wrapper = mount(<UpgradeButton update={update} upgradeURL="" />);
    const tooltip = wrapper.find(".tooltiptext");
    expect(tooltip).toExist();
    expect(tooltip.text()).toContain("New version (1.0.0) found");
  });
});
