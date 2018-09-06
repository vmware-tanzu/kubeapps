import { shallow } from "enzyme";
import * as React from "react";

import { IAppOverview, IAppState } from "../../shared/types";
import { CardGrid } from "../Card";
import AppList from "./AppList";
import AppListItem from "./AppListItem";

let defaultProps = {} as any;

beforeEach(() => {
  defaultProps = {
    apps: {} as IAppState,
    fetchApps: jest.fn(),
    filter: "",
    namespace: "default",
    pushSearchFilter: jest.fn(),
    toggleListAll: jest.fn(),
  };
});

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
          listingAll: false,
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
          listingAll: false,
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
          listingAll: false,
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
          listingAll: false,
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

it("clicking 'List All' checkbox should trigger toggleListAll", () => {
  const apps = {
    isFetching: false,
    items: [],
    listOverview: [{ releaseName: "foo" } as IAppOverview],
    listingAll: false,
  } as IAppState;
  const wrapper = shallow(
    <AppList
      {...defaultProps}
      apps={apps}
      toggleListAll={jest.fn((toggle: boolean) => {
        apps.listingAll = toggle;
      })}
    />,
  );
  const checkbox = wrapper.find('input[type="checkbox"]');
  expect(apps.listingAll).toBe(false);
  checkbox.simulate("change");
  // The last call to fetchApps should list all the apps
  const fetchCalls = defaultProps.fetchApps.mock.calls;
  expect(fetchCalls[fetchCalls.length - 1]).toEqual(["default", true]);
});
