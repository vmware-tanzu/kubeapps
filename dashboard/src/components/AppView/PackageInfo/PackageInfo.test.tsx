// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import Alert from "components/js/Alert";
import {
  AvailablePackageDetail,
  AvailablePackageReference,
  Context,
  InstalledPackageDetail,
  InstalledPackageReference,
  InstalledPackageStatus,
  InstalledPackageStatus_StatusReason,
  PackageAppVersion,
  ReconciliationOptions,
  VersionReference,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import context from "jest-plugin-context";
import { defaultStore, mountWrapper } from "shared/specs/mountWrapper";
import PackageInfo from "./PackageInfo";

const defaultProps = {
  installedPackageDetail: {
    name: "test",
    postInstallationNotes: "test",
    valuesApplied: "test",
    availablePackageRef: {
      identifier: "apache/1",
      context: { cluster: "", namespace: "package-namespace" } as Context,
      plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
    } as AvailablePackageReference,
    currentVersion: { appVersion: "10.0.0", pkgVersion: "1.0.0" } as PackageAppVersion,
    installedPackageRef: {
      identifier: "apache/1",
      pkgVersion: "1.0.0",
      context: { cluster: "", namespace: "package-namespace" } as Context,
      plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
    } as InstalledPackageReference,
    latestMatchingVersion: { appVersion: "10.0.0", pkgVersion: "1.0.0" } as PackageAppVersion,
    latestVersion: { appVersion: "10.0.0", pkgVersion: "1.0.0" } as PackageAppVersion,
    pkgVersionReference: { version: "1" } as VersionReference,
    status: {
      ready: true,
      reason: InstalledPackageStatus_StatusReason.STATUS_REASON_INSTALLED,
      userReason: "deployed",
    } as InstalledPackageStatus,
  } as InstalledPackageDetail,
  availablePackageDetail: {} as AvailablePackageDetail,
  cluster: "default",
  plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
};

it("renders an app item", () => {
  const wrapper = mountWrapper(defaultStore, <PackageInfo {...defaultProps} />);
  // Renders info about the description and versions
  const subsections = wrapper.find(".left-menu-subsection");
  expect(subsections).toHaveLength(3);
});

context("PackageUpdateInfo: when information about updates is available", () => {
  it("renders an up to date message if there are no updates", () => {
    const appWithoutUpdates = {
      ...defaultProps.installedPackageDetail,
      updateInfo: { upToDate: true },
    } as InstalledPackageDetail;
    const wrapper = mountWrapper(
      defaultStore,
      <PackageInfo {...defaultProps} installedPackageDetail={appWithoutUpdates} />,
    );
    expect(wrapper.find(".color-icon-success").text()).toContain("Up to date");
  });
  it("renders an new version found message if the package latest version is newer", () => {
    const appWithUpdates = {
      ...defaultProps.installedPackageDetail,
      latestVersion: {
        pkgVersion: "1.0.1",
        appVersion: "10.0.0",
      },
    } as InstalledPackageDetail;
    const wrapper = mountWrapper(
      defaultStore,
      <PackageInfo {...defaultProps} installedPackageDetail={appWithUpdates} />,
    );
    expect(wrapper.find(Alert).text()).toContain("A new package version is available: 1.0.1");
  });
  it("renders an new version found message if the app latest version is different", () => {
    const appWithUpdates = {
      ...defaultProps.installedPackageDetail,
      latestVersion: {
        pkgVersion: "1.0.1",
        appVersion: "10.1.0",
      },
    } as InstalledPackageDetail;
    const wrapper = mountWrapper(
      defaultStore,
      <PackageInfo {...defaultProps} installedPackageDetail={appWithUpdates} />,
    );
    expect(wrapper.find(Alert).text()).toContain("A new app version is available: 10.1.0");
  });
  it("renders an new version found message if the app latest version is different without being semver", () => {
    const appWithUpdates = {
      ...defaultProps.installedPackageDetail,
      latestVersion: {
        pkgVersion: "1.0.1",
        appVersion: "latest",
      },
    } as InstalledPackageDetail;
    const wrapper = mountWrapper(
      defaultStore,
      <PackageInfo {...defaultProps} installedPackageDetail={appWithUpdates} />,
    );
    expect(wrapper.find(Alert).text()).toContain("A new app version is available: latest");
  });
  it("renders an new version found message if the pkg latest version is different without being semver", () => {
    const appWithUpdates = {
      ...defaultProps.installedPackageDetail,
      latestVersion: {
        pkgVersion: "latest",
        appVersion: "10.0.0",
      },
    } as InstalledPackageDetail;
    const wrapper = mountWrapper(
      defaultStore,
      <PackageInfo {...defaultProps} installedPackageDetail={appWithUpdates} />,
    );
    expect(wrapper.find(Alert).text()).toContain("A new package version is available: latest");
  });
  it("renders the reconcilliation options if any", () => {
    const appWithUpdates = {
      ...defaultProps.installedPackageDetail,
      reconciliationOptions: {
        serviceAccountName: "my-sa",
        interval: "1m33s",
        suspend: false,
      } as ReconciliationOptions,
    } as InstalledPackageDetail;
    const wrapper = mountWrapper(
      defaultStore,
      <PackageInfo {...defaultProps} installedPackageDetail={appWithUpdates} />,
    );
    expect(wrapper.text()).toContain("Service Account: my-sa");
    expect(wrapper.text()).toContain("Interval: 1m33s");
  });
});
