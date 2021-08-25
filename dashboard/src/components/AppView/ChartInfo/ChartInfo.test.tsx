import Alert from "components/js/Alert";
import {
  AvailablePackageReference,
  Context,
  InstalledPackageDetail,
  InstalledPackageReference,
  InstalledPackageStatus,
  InstalledPackageStatus_StatusReason,
  PackageAppVersion,
  VersionReference,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
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
    } as AvailablePackageReference,
    currentVersion: { appVersion: "10.0.0", pkgVersion: "1.0.0" } as PackageAppVersion,
    installedPackageRef: {
      identifier: "apache/1",
      pkgVersion: "1.0.0",
      context: { cluster: "", namespace: "chart-namespace" } as Context,
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
  cluster: "default",
};

it("renders a app item", () => {
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
      updateInfo: { upToDate: false, appLatestVersion: "0.0.1", chartLatestVersion: "1.0.0" },
    } as InstalledPackageDetail;
    const wrapper = mountWrapper(
      defaultStore,
      <ChartInfo {...defaultProps} app={appWithUpdates} />,
    );
    expect(wrapper.find(Alert).text()).toContain("A new chart version is available: 1.0.0");
  });
  it("renders an new version found message if the app latest version is newer", () => {
    const appWithUpdates = {
      ...defaultProps.app,
      updateInfo: { upToDate: false, appLatestVersion: "1.1.0", chartLatestVersion: "1.0.0" },
    } as InstalledPackageDetail;
    const wrapper = mountWrapper(
      defaultStore,
      <ChartInfo {...defaultProps} app={appWithUpdates} />,
    );
    expect(wrapper.find(Alert).text()).toContain("A new app version is available: 1.1.0");
  });
  it("renders a warning if there are errors with the update info", () => {
    const appWithUpdates = {
      ...defaultProps.app,
      updateInfo: { error: new Error("Boom!"), upToDate: false, chartLatestVersion: "" },
    } as InstalledPackageDetail;
    const wrapper = mountWrapper(
      defaultStore,
      <ChartInfo {...defaultProps} app={appWithUpdates} />,
    );
    expect(wrapper.find(Alert).text()).toContain("Update check failed. Boom!");
  });
});
