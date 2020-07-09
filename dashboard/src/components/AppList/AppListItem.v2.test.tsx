import { shallow } from "enzyme";
import * as React from "react";
import { app } from "shared/url";
import InfoCard from "../InfoCard/InfoCard.v2";
import AppListItem, { IAppListItemProps } from "./AppListItem.v2";

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
    tag1Content: "Status: deployed",
    tag2Content: undefined,
    title: defaultProps.app.releaseName,
  });
});

it("should add a second label with the chart update available", () => {
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
  const wrapper = shallow(<AppListItem {...props} />);
  const card = wrapper.find(InfoCard);
  expect(card.prop("tag2Content")).toBe("New Chart: 1.1.0");
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
  const wrapper = shallow(<AppListItem {...props} />);
  const card = wrapper.find(InfoCard);
  expect(card.prop("tag2Content")).toBe("New App: 1.1.0");
});
