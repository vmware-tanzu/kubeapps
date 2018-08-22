import { shallow } from "enzyme";
import * as React from "react";
import { Link } from "react-router-dom";

import { IAppOverview, IAppState } from "../../shared/types";
import { CardGrid } from "../Card";
import { MessageAlert } from "../ErrorAlert";
import AppList from "./AppList";
import AppListItem from "./AppListItem";

it("renders a loading message if the app overview is not ready", () => {
  const wrapper = shallow(
    <AppList
      apps={{} as IAppState}
      fetchApps={jest.fn()}
      namespace="default"
      pushSearchFilter={jest.fn()}
      filter=""
    />,
  );
  expect(wrapper.text()).toBe("Loading");
});

it("renders a loading message if it's fetching apps", () => {
  const wrapper = shallow(
    <AppList
      apps={
        {
          isFetching: true,
          items: [],
          listOverview: [],
        } as IAppState
      }
      fetchApps={jest.fn()}
      namespace="default"
      pushSearchFilter={jest.fn()}
      filter=""
    />,
  );
  expect(wrapper.text()).toContain("Loading");
});

it("renders a welcome message if no apps are available", () => {
  const wrapper = shallow(
    <AppList
      apps={
        {
          isFetching: false,
          items: [],
          listOverview: [],
        } as IAppState
      }
      fetchApps={jest.fn()}
      namespace="default"
      pushSearchFilter={jest.fn()}
      filter=""
    />,
  );
  expect(
    wrapper
      .find(MessageAlert)
      .children()
      .text(),
  ).toContain("Deploy applications on your Kubernetes cluster with a single click");
  expect(
    wrapper
      .find(MessageAlert)
      .children()
      .find(Link)
      .props(),
  ).toMatchObject({ to: `/charts`, children: "Deploy App" });
});

it("renders a CardGrid with the available Apps", () => {
  const wrapper = shallow(
    <AppList
      apps={
        {
          isFetching: false,
          items: [],
          listOverview: [
            {
              releaseName: "foo",
            } as IAppOverview,
          ],
        } as IAppState
      }
      fetchApps={jest.fn()}
      namespace="default"
      pushSearchFilter={jest.fn()}
      filter=""
    />,
  );
  expect(
    wrapper
      .find(CardGrid)
      .children()
      .find(AppListItem)
      .key(),
  ).toBe("foo");
});

it("filters apps", () => {
  const wrapper = shallow(
    <AppList
      apps={
        {
          isFetching: false,
          items: [],
          listOverview: [
            {
              releaseName: "foo",
            } as IAppOverview,
            {
              releaseName: "bar",
            } as IAppOverview,
          ],
        } as IAppState
      }
      fetchApps={jest.fn()}
      namespace="default"
      pushSearchFilter={jest.fn()}
      filter="bar"
    />,
  );
  expect(
    wrapper
      .find(CardGrid)
      .children()
      .find(AppListItem)
      .key(),
  ).toBe("bar");
});
