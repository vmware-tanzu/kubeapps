import { shallow } from "enzyme";
import * as React from "react";

import { IAppOverview, IAppState } from "../../shared/types";
import { CardGrid } from "../Card";
import AppList from "./AppList";
import AppListItem from "./AppListItem";

const defaultProps = {
  apps: {} as IAppState,
  fetchApps: jest.fn(),
  filter: "",
  namespace: "default",
  pushSearchFilter: jest.fn(),
};

it("renders a loading message if apps object is empty", () => {
  const wrapper = shallow(<AppList {...defaultProps} />);
  expect(wrapper.text()).toBe("Loading");
});

it("renders a loading message if it's fetching apps", () => {
  const wrapper = shallow(
    <AppList
      {...defaultProps}
      apps={
        {
          isFetching: true,
          items: [],
          listOverview: [],
        } as IAppState
      }
    />,
  );
  expect(wrapper.text()).toContain("Loading");
});

it("renders a welcome message if no apps are available", () => {
  const wrapper = shallow(
    <AppList
      {...defaultProps}
      apps={
        {
          isFetching: false,
          items: [],
          listOverview: [],
        } as IAppState
      }
    />,
  );
  expect(wrapper).toMatchSnapshot();
});

it("renders a CardGrid with the available Apps", () => {
  const wrapper = shallow(
    <AppList
      {...defaultProps}
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
      {...defaultProps}
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
