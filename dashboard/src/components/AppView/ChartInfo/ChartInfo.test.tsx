import Alert from "components/js/Alert";
import context from "jest-plugin-context";

import { hapi } from "shared/hapi/release";
import { defaultStore, mountWrapper } from "shared/specs/mountWrapper";
import { IRelease } from "shared/types";
import ChartInfo from "./ChartInfo";

const defaultProps = {
  app: {
    chart: {
      metadata: {
        name: "bar",
        appVersion: "0.0.1",
        description: "test chart",
        icon: "icon.png",
        version: "1.0.0",
      },
    },
    name: "foo",
  } as hapi.release.Release,
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
    const appWithoutUpdates = { ...defaultProps.app, updateInfo: { upToDate: true } } as IRelease;
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
    } as IRelease;
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
    } as IRelease;
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
    } as IRelease;
    const wrapper = mountWrapper(
      defaultStore,
      <ChartInfo {...defaultProps} app={appWithUpdates} />,
    );
    expect(wrapper.find(Alert).text()).toContain("Update check failed. Boom!");
  });
});
