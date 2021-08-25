import Tooltip from "components/js/Tooltip";
import { shallow } from "enzyme";
import {
  Context,
  InstalledPackageReference,
  InstalledPackageStatus,
  InstalledPackageStatus_StatusReason,
  InstalledPackageSummary,
  PackageAppVersion,
  VersionReference,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { defaultStore, mountWrapper } from "shared/specs/mountWrapper";
import { app } from "shared/url";
import InfoCard from "../InfoCard/InfoCard";
import AppListItem, { IAppListItemProps } from "./AppListItem";

const defaultProps = {
  app: {
    name: "foo",
    installedPackageRef: {
      identifier: "apache/1",
      pkgVersion: "1.0.0",
      context: { cluster: "", namespace: "chart-namespace" } as Context,
    } as InstalledPackageReference,
    status: {
      ready: true,
      reason: InstalledPackageStatus_StatusReason.STATUS_REASON_INSTALLED,
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
  const wrapper = shallow(<AppListItem {...defaultProps} />);
  const card = wrapper.find(InfoCard);
  expect(card.props()).toMatchObject({
    description: defaultProps.app.shortDescription,
    icon: "placeholder.png",
    link: app.apps.get(
      defaultProps.cluster,
      defaultProps.app.installedPackageRef?.context?.namespace!,
      defaultProps.app.name,
    ),
    tag1Class: "label-success",
    tag1Content: "deployed",
    title: defaultProps.app.name,
  });
});

it("should add a tooltip with the chart update available", () => {
  const props = {
    ...defaultProps,
    app: {
      ...defaultProps.app,
      latestVersion: { appVersion: "1.0.0", pkgVersion: "1.1.0" } as PackageAppVersion,
      currentVersion: { appVersion: "1.0.0", pkgVersion: "1.0.0" } as PackageAppVersion,
    },
  } as IAppListItemProps;
  const wrapper = mountWrapper(defaultStore, <AppListItem {...props} />);
  const tooltip = wrapper.find(Tooltip);
  expect(tooltip.text()).toBe("A new package version is available: 1.1.0");
});

it("should add a second label with the app update available", () => {
  const props = {
    ...defaultProps,
    app: {
      ...defaultProps.app,
      latestVersion: { appVersion: "1.1.0", pkgVersion: "1.1.0" } as PackageAppVersion,
      currentVersion: { appVersion: "1.0.0", pkgVersion: "1.0.0" } as PackageAppVersion,
    },
  } as IAppListItemProps;
  const wrapper = mountWrapper(defaultStore, <AppListItem {...props} />);
  const tooltip = wrapper.find(Tooltip);
  expect(tooltip.text()).toBe("A new app version is available: 1.1.0");
});

// TODO(agamez): Test temporarily commented out
// it("doesn't include a double v prefix", () => {
//   const props = {
//     ...defaultProps,
//     app: {
//       ...defaultProps.app,
//       latestVersion: { appVersion: "1.0.0", pkgVersion: "1.1.0" } as PackageAppVersion,
//       currentVersion: { appVersion: "1.0.0", pkgVersion: "1.0.0" } as PackageAppVersion,
//     },
//   } as IAppListItemProps;
//   const wrapper = mountWrapper(defaultStore, <AppListItem {...props} />);
//   expect(wrapper.find("span").findWhere(s => s.text() === "App: foo v1.0.0")).toExist();
// });
