import Tooltip from "components/js/Tooltip";
import { shallow } from "enzyme";
import { defaultStore, mountWrapper } from "shared/specs/mountWrapper";
import { app } from "shared/url";
import InfoCard from "../InfoCard/InfoCard";
import AppListItem, { IAppListItemProps } from "./AppListItem";

const defaultProps = {
  app: {
    namespace: "default",
    releaseName: "foo",
    status: "DEPLOYED",
    version: "1.0.0",
    chart: "myapp",
    chartMetadata: {
      appVersion: "1.0.0",
      description: "this is a description",
    },
  },
  cluster: "default",
} as IAppListItemProps;

it("renders an app item", () => {
  const wrapper = shallow(<AppListItem {...defaultProps} />);
  const card = wrapper.find(InfoCard);
  expect(card.props()).toMatchObject({
    description: defaultProps.app.chartMetadata.description,
    icon: "placeholder.png",
    link: app.apps.get(
      defaultProps.cluster,
      defaultProps.app.namespace,
      defaultProps.app.releaseName,
    ),
    tag1Class: "label-success",
    tag1Content: "deployed",
    title: defaultProps.app.releaseName,
  });
});

it("should add a tooltip with the chart update available", () => {
  const props = {
    ...defaultProps,
    app: {
      ...defaultProps.app,
      chartMetadata: {
        appVersion: "1.1.0",
      },
      updateInfo: {
        upToDate: false,
        chartLatestVersion: "1.1.0",
        appLatestVersion: "1.1.0",
        repository: { name: "", url: "" },
      },
    },
  } as IAppListItemProps;
  const wrapper = mountWrapper(defaultStore, <AppListItem {...props} />);
  const tooltip = wrapper.find(Tooltip);
  expect(tooltip.text()).toBe("New Chart Version: 1.1.0");
});

it("should add a second label with the app update available", () => {
  const props = {
    ...defaultProps,
    app: {
      ...defaultProps.app,
      chartMetadata: {
        appVersion: "1.0.0",
      },
      updateInfo: {
        upToDate: false,
        chartLatestVersion: "1.0.0",
        appLatestVersion: "1.1.0",
        repository: { name: "", url: "" },
      },
    },
  } as IAppListItemProps;
  const wrapper = mountWrapper(defaultStore, <AppListItem {...props} />);
  const tooltip = wrapper.find(Tooltip);
  expect(tooltip.text()).toBe("New App Version: 1.1.0");
});

it("doesn't include a double v prefix", () => {
  const props = {
    ...defaultProps,
    app: {
      ...defaultProps.app,
      chartMetadata: {
        name: "foo",
        appVersion: "v1.0.0",
      },
      updateInfo: {},
    },
  } as IAppListItemProps;
  const wrapper = mountWrapper(defaultStore, <AppListItem {...props} />);
  expect(wrapper.find("span").findWhere(s => s.text() === "App: foo v1.0.0")).toExist();
});
