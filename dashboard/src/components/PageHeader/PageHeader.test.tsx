// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { mount } from "enzyme";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import PageHeader from "./PageHeader";

const defaultProps = {
  title: "This is a title!",
};

it("should renders a h1 title", () => {
  const wrapper = mount(<PageHeader {...defaultProps} />);
  expect(wrapper.find("h1")).toIncludeText(defaultProps.title);
});

it("should render a smaller title", () => {
  const wrapper = mount(<PageHeader {...defaultProps} titleSize="md" />);
  expect(wrapper.find("h3")).toIncludeText(defaultProps.title);
});

it("includes an icon", () => {
  const wrapper = mount(<PageHeader {...defaultProps} icon="icon.png" />);
  expect(wrapper.find("img").prop("src")).toBe("icon.png");
});

it("includes subtitle", () => {
  const wrapper = mount(<PageHeader {...defaultProps} subtitle={<span id="test">test</span>} />);
  expect(wrapper.find("span").prop("id")).toBe("test");
});

it("includes a filter component", () => {
  const wrapper = mount(<PageHeader {...defaultProps} filter={<div id="foo" />} />);
  expect(wrapper.find("#foo")).toExist();
});

it("renders a Helm subtitle", () => {
  const wrapper = mount(
    <PageHeader {...defaultProps} plugin={{ name: "helm.packages", version: "0.0.1" } as Plugin} />,
  );
  expect(wrapper.find("img").prop("src")).toBe("helm.svg");
  expect(wrapper.text()).toContain("Helm Chart");
});

it("renders a Flux subtitle", () => {
  const wrapper = mount(
    <PageHeader
      {...defaultProps}
      plugin={{ name: "fluxv2.packages", version: "0.0.1" } as Plugin}
    />,
  );
  expect(wrapper.find("img").prop("src")).toBe("flux.svg");
  expect(wrapper.text()).toContain("Helm Chart");
});

it("renders a Carvel subtitle", () => {
  const wrapper = mount(
    <PageHeader
      {...defaultProps}
      plugin={{ name: "kapp_controller.packages", version: "0.0.1" } as Plugin}
    />,
  );
  expect(wrapper.find("img").prop("src")).toBe("carvel.svg");
  expect(wrapper.text()).toContain("Carvel Package");
});

it("renders an Operator subtitle", () => {
  const wrapper = mount(<PageHeader {...defaultProps} operator={true} />);
  expect(wrapper.find("img").prop("src")).toBe("olm-icon.svg");
  expect(wrapper.text()).toContain("Operator");
});

it("renders a version section", () => {
  const wrapper = mount(<PageHeader {...defaultProps} version={<div id="foo" />} />);
  expect(wrapper.find("#foo")).toExist();
});

it("renders buttons section", () => {
  const wrapper = mount(
    <PageHeader
      {...defaultProps}
      buttons={[<button key="foo" id="foo" />, <button key="bar" id="bar" />]}
    />,
  );
  expect(wrapper.find("#foo")).toExist();
  expect(wrapper.find("#bar")).toExist();
});
