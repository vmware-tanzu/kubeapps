// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { mount } from "enzyme";
import {
  AvailablePackageDetail,
  Context,
  PackageAppVersion,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import PackageHeader, { IPackageHeaderProps } from "./PackageHeader";
const testProps: IPackageHeaderProps = {
  availablePackageDetail: {
    shortDescription: "A Test Package",
    name: "foo",
    categories: [""],
    displayName: "foo",
    iconUrl: "api/test.jpg",
    repoUrl: "",
    homeUrl: "",
    sourceUrls: [],
    longDescription: "",
    availablePackageRef: {
      identifier: "testrepo/foo",
      context: { cluster: "default", namespace: "kubeapps" } as Context,
      plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
    },
    valuesSchema: "",
    defaultValues: "",
    maintainers: [],
    readme: "",
    version: {
      pkgVersion: "1.2.3",
      appVersion: "4.5.6",
    },
  } as AvailablePackageDetail,
  versions: [
    {
      pkgVersion: "0.1.2",
      appVersion: "1.2.3",
    },
  ] as PackageAppVersion[],
  onSelect: jest.fn(),
};

it("renders a header for the package with display name", () => {
  const wrapper = mount(<PackageHeader {...testProps} />);
  expect(wrapper.text()).toContain("foo");
  expect(wrapper.text()).not.toContain("testrepo/foo");
});

it("displays the appVersion", () => {
  const wrapper = mount(<PackageHeader {...testProps} />);
  expect(wrapper.text()).toContain("1.2.3");
});

it("uses the icon", () => {
  const wrapper = mount(<PackageHeader {...testProps} />);
  const icon = wrapper.find("img").filterWhere(i => i.prop("alt") === "icon");
  expect(icon.exists()).toBe(true);
  expect(icon.props()).toMatchObject({ src: "api/test.jpg" });
});

it("uses the first version as default in the select input", () => {
  const versions: PackageAppVersion[] = [
    {
      pkgVersion: "1.2.3",
      appVersion: "10.0.0",
    },
    {
      pkgVersion: "1.2.4",
      appVersion: "10.0.0",
    },
  ];
  const wrapper = mount(<PackageHeader {...testProps} versions={versions} />);
  expect(wrapper.find("select").prop("value")).toBe("1.2.3");
});

it("uses the current version as default in the select input", () => {
  const versions: PackageAppVersion[] = [
    {
      pkgVersion: "1.2.3",
      appVersion: "10.0.0",
    },
    {
      pkgVersion: "1.2.4",
      appVersion: "10.0.0",
    },
  ];
  const wrapper = mount(
    <PackageHeader {...testProps} versions={versions} currentVersion="1.2.4" />,
  );
  expect(wrapper.find("select").prop("value")).toBe("1.2.4");
});
