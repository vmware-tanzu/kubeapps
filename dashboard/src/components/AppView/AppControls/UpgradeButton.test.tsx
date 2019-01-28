import { mount, shallow } from "enzyme";
import context from "jest-plugin-context";
import * as React from "react";
import { Redirect } from "react-router";

import UpgradeButton from "./UpgradeButton";

it("renders a redirect when clicking upgrade", () => {
  const push = jest.fn();
  const wrapper = shallow(
    <UpgradeButton releaseName="foo" releaseNamespace="default" push={push} />,
  );
  const button = wrapper.find(".button").filterWhere(i => i.text() === "Upgrade");
  expect(button.exists()).toBe(true);
  expect(wrapper.find(Redirect).exists()).toBe(false);

  button.simulate("click");
  expect(push.mock.calls.length).toBe(1);
  expect(push.mock.calls[0]).toEqual(["/apps/ns/default/upgrade/foo"]);
});

context("when a new version is available", () => {
  it("should show a tooltip to notify the new version", () => {
    const wrapper = mount(
      <UpgradeButton updateVersion="1.0.0" releaseName="" releaseNamespace="" push={jest.fn()} />,
    );
    const tooltip = wrapper.find(".tooltiptext");
    expect(tooltip).toExist();
    expect(tooltip.text()).toContain("New version (1.0.0) found");
  });
});
