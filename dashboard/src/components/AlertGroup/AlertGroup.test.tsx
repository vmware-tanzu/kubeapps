// Copyright 2021-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsAlert, CdsAlertGroup } from "@cds/react/alert";
import { act } from "@testing-library/react";
import { mount } from "enzyme";
import AlertGroup from "./AlertGroup";

it("should render children components", () => {
  const wrapper = mount(<AlertGroup>foo</AlertGroup>);
  expect(wrapper.text()).toContain("foo");
});

it("should render alertActions", () => {
  const wrapper = mount(<AlertGroup alertActions={<div>bar</div>}>foo</AlertGroup>);
  expect(wrapper.text()).toContain("foo");
  expect(wrapper.text()).toContain("bar");
});

it("should close the alert", () => {
  const wrapper = mount(<AlertGroup closable={true}>foo</AlertGroup>);
  act(() => {
    (wrapper.find(CdsAlert).prop("onCloseChange") as any)();
  });
  wrapper.update();
  expect(wrapper.find(CdsAlertGroup).prop("hidden")).toBe(true);
});

it("should set custom properties", () => {
  const customProps = {
    status: "info",
    type: "flat",
    size: "sm",
  };
  const wrapper = mount(
    <AlertGroup closable={true} {...(customProps as any)}>
      foo
    </AlertGroup>,
  );
  expect(wrapper.find(CdsAlertGroup).props()).toMatchObject(customProps);
});
