import context from "jest-plugin-context";
import * as React from "react";
import { ArrowUpCircle } from "react-feather";
import { Redirect } from "react-router";
import { getStore, mountWrapper } from "shared/specs/mountWrapper";
import * as url from "shared/url";
import UpgradeButton, { IUpgradeButtonProps } from "./UpgradeButton";

const defaultProps = {
  releaseName: "foo",
  releaseNamespace: "default",
  push: jest.fn(),
} as IUpgradeButtonProps;

it("renders a redirect when clicking upgrade", () => {
  const push = jest.fn();
  const store = getStore({});
  const props = { ...defaultProps, push };
  const wrapper = mountWrapper(store, <UpgradeButton {...props} />);
  const button = wrapper.find(".button").filterWhere(i => i.text() === "Upgrade");
  expect(button.exists()).toBe(true);
  expect(wrapper.find(Redirect).exists()).toBe(false);

  button.simulate("click");
  expect(push.mock.calls.length).toBe(1);
  expect(push.mock.calls[0]).toEqual([url.app.apps.upgrade("default-cluster", "default", "foo")]);
});

context("when a new version is available", () => {
  it("should show a modify the style", () => {
    const store = getStore({});
    const props = { ...defaultProps, newVersion: true };
    const wrapper = mountWrapper(store, <UpgradeButton {...props} />);
    const icon = wrapper.find(ArrowUpCircle);
    expect(icon).toExist();
  });
});
