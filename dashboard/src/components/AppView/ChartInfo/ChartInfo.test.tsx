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
  VersionReference,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import context from "jest-plugin-context";
import { defaultStore, mountWrapper } from "shared/specs/mountWrapper";
import ChartInfo from "./ChartInfo";

const defaultProps = {
  app: {
    name: "test",
    postInstallationNotes: "test",
    valuesApplied: "test",
    availablePackageRef: {
      identifier: "apache/1",
      context: { cluster: "", namespace: "chart-namespace" } as Context,
      plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
    } as AvailablePackageReference,
    currentVersion: { appVersion: "10.0.0", pkgVersion: "1.0.0" } as PackageAppVersion,
    installedPackageRef: {
      identifier: "apache/1",
      pkgVersion: "1.0.0",
      context: { cluster: "", namespace: "chart-namespace" } as Context,
      plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
    } as InstalledPackageReference,
    latestMatchingVersion: { appVersion: "10.0.0", pkgVersion: "1.0.0" } as PackageAppVersion,
    latestVersion: { appVersion: "10.0.0", pkgVersion: "1.0.0" } as PackageAppVersion,
    pkgVersionReference: { version: "1" } as VersionReference,
    reconciliationOptions: {},
    status: {
      ready: true,
      reason: InstalledPackageStatus_StatusReason.STATUS_REASON_INSTALLED,
      userReason: "deployed",
    } as InstalledPackageStatus,
  } as InstalledPackageDetail,
  appDetails: {} as AvailablePackageDetail,
  cluster: "default",
  plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
};

it("renders an app item", () => {
  const wrapper = mountWrapper(defaultStore, <ChartInfo {...defaultProps} />);
  // Renders info about the description and versions
  const subsections = wrapper.find(".left-menu-subsection");
  expect(subsections).toHaveLength(2);
});

context("ChartUpdateInfo: when information about updates is available", () => {
  it("renders an up to date message if there are no updates", () => {
    const appWithoutUpdates = {
      ...defaultProps.app,
      updateInfo: { upToDate: true },
    } as InstalledPackageDetail;
    const wrapper = mountWrapper(
      defaultStore,
      <ChartInfo {...defaultProps} app={appWithoutUpdates} />,
    );
    expect(wrapper.find(".color-icon-success").text()).toContain("Up to date");
  });
  it("renders an new version found message if the chart latest version is newer", () => {
    const appWithUpdates = {
      ...defaultProps.app,
      latestVersion: {
        pkgVersion: "1.0.1",
        appVersion: "10.0.0",
      },
    } as InstalledPackageDetail;
    const wrapper = mountWrapper(
      defaultStore,
      <ChartInfo {...defaultProps} app={appWithUpdates} />,
    );
    expect(wrapper.find(Alert).text()).toContain("A new package version is available: 1.0.1");
  });
  it("renders an new version found message if the app latest version is different", () => {
    const appWithUpdates = {
      ...defaultProps.app,
      latestVersion: {
        pkgVersion: "1.0.1",
        appVersion: "10.1.0",
      },
    } as InstalledPackageDetail;
    const wrapper = mountWrapper(
      defaultStore,
      <ChartInfo {...defaultProps} app={appWithUpdates} />,
    );
    expect(wrapper.find(Alert).text()).toContain("A new app version is available: 10.1.0");
  });
  it("renders an new version found message if the app latest version is different without being semver", () => {
    const appWithUpdates = {
      ...defaultProps.app,
      latestVersion: {
        pkgVersion: "1.0.1",
        appVersion: "latest",
      },
    } as InstalledPackageDetail;
    const wrapper = mountWrapper(
      defaultStore,
      <ChartInfo {...defaultProps} app={appWithUpdates} />,
    );
    expect(wrapper.find(Alert).text()).toContain("A new app version is available: latest");
  });
  it("renders an new version found message if the pkg latest version is different without being semver", () => {
    const appWithUpdates = {
      ...defaultProps.app,
      latestVersion: {
        pkgVersion: "latest",
        appVersion: "10.0.0",
      },
    } as InstalledPackageDetail;
    const wrapper = mountWrapper(
      defaultStore,
      <ChartInfo {...defaultProps} app={appWithUpdates} />,
    );
    expect(wrapper.find(Alert).text()).toContain("A new package version is available: latest");
  });
});
