// Copyright 2018-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { act } from "@testing-library/react";
import {
  Context,
  InstalledPackageReference,
  InstalledPackageStatus,
  InstalledPackageStatus_StatusReason,
  InstalledPackageSummary,
  PackageAppVersion,
  VersionReference,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages_pb";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins_pb";
import { Tooltip } from "react-tooltip";
import { defaultStore, mountWrapper } from "shared/specs/mountWrapper";
import { app } from "shared/url";
import InfoCard from "../InfoCard/InfoCard";
import AppListItem, { IAppListItemProps } from "./AppListItem";

const defaultProps = {
  app: {
    name: "foo",
    pkgDisplayName: "foo",
    installedPackageRef: new InstalledPackageReference({
      identifier: "foo",
      context: { cluster: "default", namespace: "package-namespace" } as Context,
      plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
    }),
    status: {
      ready: true,
      reason: InstalledPackageStatus_StatusReason.INSTALLED,
      userReason: "deployed",
    } as InstalledPackageStatus,
    latestMatchingVersion: { appVersion: "10.0.0", pkgVersion: "1.0.0" } as PackageAppVersion,
    latestVersion: { appVersion: "10.0.0", pkgVersion: "1.0.0" } as PackageAppVersion,
    currentVersion: { appVersion: "0.1.0", pkgVersion: "1.0.0" } as PackageAppVersion,
    pkgVersionReference: { version: "1" } as VersionReference,
  } as InstalledPackageSummary,
  cluster: "default",
} as IAppListItemProps;

it("renders an app item", () => {
  const wrapper = mountWrapper(defaultStore, <AppListItem {...defaultProps} />);
  const card = wrapper.find(InfoCard);
  expect(card.props()).toMatchObject({
    description: defaultProps.app.shortDescription,
    icon: "placeholder.svg",
    link: app.apps.get({
      context: {
        cluster: defaultProps.cluster,
        namespace: defaultProps.app.installedPackageRef?.context?.namespace ?? "",
      },
      identifier: defaultProps.app.name,
      plugin: { name: "my.plugin", version: "0.0.1" },
    } as InstalledPackageReference),
    tag1Class: "label-success",
    tag1Content: "installed",
    tag2Class: "label-info-secondary",
    tag2Content: "my.plugin",
    title: defaultProps.app.name,
  });
});

it("should add a tooltip with the package update available", async () => {
  const props = {
    ...defaultProps,
    app: {
      ...defaultProps.app,
      latestVersion: { appVersion: "1.0.0", pkgVersion: "1.1.0" } as PackageAppVersion,
      currentVersion: { appVersion: "1.0.0", pkgVersion: "1.0.0" } as PackageAppVersion,
    },
  } as IAppListItemProps;
  const wrapper = mountWrapper(defaultStore, <AppListItem {...props} />);

  // Cloning the tooltip with the isOpen prop set to true,
  // this way we can later test the tooltip content
  act(() => {
    wrapper.setProps({
      children: (
        <Tooltip {...wrapper.find(Tooltip).props()} isOpen={true}>
          {wrapper.find(Tooltip).prop("children")}
        </Tooltip>
      ),
    });
  });
  wrapper.update();

  expect(wrapper.text()).toBe("A new package version is available: 1.1.0");
});

it("should add a tooltip with the app update available", () => {
  const props = {
    ...defaultProps,
    app: {
      ...defaultProps.app,
      latestVersion: { appVersion: "1.1.0", pkgVersion: "1.1.0" } as PackageAppVersion,
      currentVersion: { appVersion: "1.0.0", pkgVersion: "1.0.0" } as PackageAppVersion,
    },
  } as IAppListItemProps;
  const wrapper = mountWrapper(defaultStore, <AppListItem {...props} />);

  // Cloning the tooltip with the isOpen prop set to true,
  // this way we can later test the tooltip content
  act(() => {
    wrapper.setProps({
      children: (
        <Tooltip {...wrapper.find(Tooltip).props()} isOpen={true}>
          {wrapper.find(Tooltip).prop("children")}
        </Tooltip>
      ),
    });
  });
  wrapper.update();

  expect(wrapper.text()).toBe("A new app version is available: 1.1.0");
});

it("should add a tooltip with the app update available without requiring semver versioning", () => {
  // The AppVersion is not always required to be semver2, certainly not in Helm's Chart.yaml.
  const props = {
    ...defaultProps,
    app: {
      ...defaultProps.app,
      latestVersion: { appVersion: "latest-crack", pkgVersion: "1.1.0" } as PackageAppVersion,
      currentVersion: { appVersion: "1.0.0", pkgVersion: "1.0.0" } as PackageAppVersion,
    },
  } as IAppListItemProps;
  const wrapper = mountWrapper(defaultStore, <AppListItem {...props} />);

  // Cloning the tooltip with the isOpen prop set to true,
  // this way we can later test the tooltip content
  act(() => {
    wrapper.setProps({
      children: (
        <Tooltip {...wrapper.find(Tooltip).props()} isOpen={true}>
          {wrapper.find(Tooltip).prop("children")}
        </Tooltip>
      ),
    });
  });
  wrapper.update();

  expect(wrapper.text()).toBe("A new app version is available: latest-crack");
});

it("should add a tooltip with the pkg update available without requiring semver versioning", () => {
  const props = {
    ...defaultProps,
    app: {
      ...defaultProps.app,
      latestVersion: { appVersion: "1.0.0", pkgVersion: "latest" } as PackageAppVersion,
      currentVersion: { appVersion: "1.0.0", pkgVersion: "1.0.0" } as PackageAppVersion,
    },
  } as IAppListItemProps;
  const wrapper = mountWrapper(defaultStore, <AppListItem {...props} />);

  // Cloning the tooltip with the isOpen prop set to true,
  // this way we can later test the tooltip content
  act(() => {
    wrapper.setProps({
      children: (
        <Tooltip {...wrapper.find(Tooltip).props()} isOpen={true}>
          {wrapper.find(Tooltip).prop("children")}
        </Tooltip>
      ),
    });
  });
  wrapper.update();

  expect(wrapper.text()).toBe("A new package version is available: latest");
});

it("doesn't include a double v prefix", () => {
  const props = {
    ...defaultProps,
    app: {
      ...defaultProps.app,
      latestVersion: { appVersion: "1.0.0", pkgVersion: "1.1.0" } as PackageAppVersion,
      currentVersion: { appVersion: "1.0.0", pkgVersion: "1.0.0" } as PackageAppVersion,
    },
  } as IAppListItemProps;
  const wrapper = mountWrapper(defaultStore, <AppListItem {...props} />);
  expect(wrapper.find("span").findWhere(s => s.text() === "App: foo v1.0.0")).toExist();
});

it("includes namespace", () => {
  const props = {
    ...defaultProps,
    app: {
      ...defaultProps.app,
      latestVersion: { appVersion: "1.0.0", pkgVersion: "1.1.0" } as PackageAppVersion,
      currentVersion: { appVersion: "1.0.0", pkgVersion: "1.0.0" } as PackageAppVersion,
    },
  } as IAppListItemProps;
  const wrapper = mountWrapper(defaultStore, <AppListItem {...props} />);
  expect(
    wrapper.find("span").findWhere(s => s.text() === "Namespace: package-namespace"),
  ).toExist();
});
