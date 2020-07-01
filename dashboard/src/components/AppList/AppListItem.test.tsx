import { shallow } from "enzyme";
import * as React from "react";
import { Link } from "react-router-dom";
import * as url from "../../shared/url";
import InfoCard from "../InfoCard";
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
    },
  },
  cluster: "default",
} as IAppListItemProps;

it("renders an app item", () => {
  const wrapper = shallow(<AppListItem {...defaultProps} />);
  const card = wrapper.find(InfoCard).shallow();
  expect(
    card
      .find(Link)
      .at(0)
      .props().title,
  ).toBe("foo");
  expect(
    card
      .find(Link)
      .at(0)
      .props().to,
  ).toBe(url.app.apps.get("default", "default", "foo"));
  expect(card.find(".type-color-light-blue").text()).toBe("myapp v1.0.0");
  expect(card.find(".deployed").exists()).toBe(true);
  expect(card.find(".ListItem__content__info_tag-1").text()).toBe("default");
  expect(card.find(".ListItem__content__info_tag-2").text()).toBe("deployed");
});

it("should set a banner if there are chart updates available", () => {
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
  expect(card.prop("banner")).toBe("Update available");
});

it("should set a banner if there are app updates available", () => {
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
  expect(card.prop("banner")).toBe("Update available");
});

it("should not set a banner if there are errors in the update info", () => {
  const props = {
    ...defaultProps,
    app: {
      ...defaultProps.app,
      chartMetadata: {
        appVersion: "1.0.0",
      },
      updateInfo: {
        upToDate: false,
        error: new Error("Boom!"),
        chartLatestVersion: "1.0.0",
        appLatestVersion: "1.1.0",
        repository: { name: "", url: "" },
      },
    },
  } as IAppListItemProps;
  const wrapper = shallow(<AppListItem {...props} />);
  const card = wrapper.find(InfoCard);
  expect(card.prop("banner")).toBe(undefined);
});
