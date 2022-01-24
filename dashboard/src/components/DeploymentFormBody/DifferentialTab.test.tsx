// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsIcon } from "@cds/react/icon";
import { mount } from "enzyme";
import { act } from "react-dom/test-utils";
import DifferentialTab from "./DifferentialTab";

describe("when installing", () => {
  it("should show the changes icon if the values change", () => {
    const wrapper = mount(
      <DifferentialTab
        deploymentEvent="install"
        deployedValues=""
        defaultValues="foo"
        appValues="bar"
      />,
    );
    expect(wrapper.find(CdsIcon).prop("hidden")).toBe(false);
  });

  it("should hide the changes icon if the values are the same", () => {
    const wrapper = mount(
      <DifferentialTab
        deploymentEvent="install"
        deployedValues=""
        defaultValues="foo"
        appValues="foo"
      />,
    );
    expect(wrapper.find(CdsIcon).prop("hidden")).toBe(true);
  });

  it("clicking the tab removes the icon", () => {
    const wrapper = mount(
      <DifferentialTab
        deploymentEvent="install"
        deployedValues=""
        defaultValues="foo"
        appValues="bar"
      />,
    );
    expect(wrapper.find(CdsIcon).prop("hidden")).toBe(false);
    act(() => {
      wrapper.find("div").simulate("click");
    });
    wrapper.update();
    expect(wrapper.find(CdsIcon).prop("hidden")).toBe(true);
  });

  it("setting default values removes the icon", () => {
    const wrapper = mount(
      <DifferentialTab
        deploymentEvent="install"
        deployedValues=""
        defaultValues="foo"
        appValues="bar"
      />,
    );
    expect(wrapper.find(CdsIcon).prop("hidden")).toBe(false);
    act(() => {
      wrapper.setProps({ appValues: "foo" });
    });
    wrapper.update();
    expect(wrapper.find(CdsIcon).prop("hidden")).toBe(true);
  });
});

describe("when upgrading", () => {
  it("should show the changes icon if the values change", () => {
    const wrapper = mount(
      <DifferentialTab
        deploymentEvent="upgrade"
        deployedValues="foo"
        defaultValues="foo"
        appValues="bar"
      />,
    );
    expect(wrapper.find(CdsIcon).prop("hidden")).toBe(false);
  });

  it("should hide the changes icon if the values are the same", () => {
    const wrapper = mount(
      <DifferentialTab
        deploymentEvent="upgrade"
        deployedValues="foo"
        defaultValues="foo"
        appValues="foo"
      />,
    );
    expect(wrapper.find(CdsIcon).prop("hidden")).toBe(true);
  });
});
