import { shallow } from "enzyme";
import context from "jest-plugin-context";
import * as React from "react";

import itBehavesLike from "../../shared/specs";
import { IAppOverview, IAppState } from "../../shared/types";
import { CardGrid } from "../Card";
import { ErrorSelector } from "../ErrorAlert";
import { genericMessage } from "../ErrorAlert/UnexpectedErrorAlert";
import AppList from "./AppList";
import AppListItem from "./AppListItem";

let props = {} as any;

const defaultProps: any = {
  apps: {} as IAppState,
  fetchApps: jest.fn(),
  filter: "",
  namespace: "default",
  pushSearchFilter: jest.fn(),
  toggleListAll: jest.fn(),
  fetchAppsWithUpdateInfo: jest.fn(),
};

context("while fetching apps", () => {
  props = { ...defaultProps, apps: { isFetching: true } };

  itBehavesLike("aLoadingComponent", { component: AppList, props });

  it("matches the snapshot", () => {
    const wrapper = shallow(<AppList {...props} />);
    expect(wrapper).toMatchSnapshot();
  });

  it("renders a Application header", () => {
    const wrapper = shallow(<AppList {...props} />);
    expect(wrapper.find("h1").text()).toContain("Applications");
  });

  it("shows the search filter, show deleted checkbox and deploy button", () => {
    const wrapper = shallow(<AppList {...props} />);
    expect(wrapper.find("SearchFilter")).toExist();
    expect(wrapper.find("label.checkbox").key()).toEqual("listall");
    expect(wrapper.find(".deploy-button")).toExist();
  });
});

context("when fetched but not apps available", () => {
  beforeEach(() => {
    props = { ...defaultProps, apps: { listOverview: [] } };
  });

  it("matches the snapshot", () => {
    const wrapper = shallow(<AppList {...props} />);
    expect(wrapper).toMatchSnapshot();
  });

  it("renders a welcome message", () => {
    const wrapper = shallow(<AppList {...props} />);
    expect(wrapper.find("MessageAlertPage")).toExist();
  });

  it("shows the search filter, show deleted checkbox and deploy button", () => {
    const wrapper = shallow(<AppList {...props} />);
    expect(wrapper.find("SearchFilter")).toExist();
    expect(wrapper.find("label.checkbox").key()).toEqual("listall");
    expect(wrapper.find(".deploy-button")).toExist();
  });
});

context("when error present", () => {
  beforeEach(() => {
    props = { ...defaultProps, apps: { error: "Boom!" } };
  });

  it("matches the snapshot", () => {
    const wrapper = shallow(<AppList {...props} />);
    expect(wrapper).toMatchSnapshot();
  });

  it("renders a generic error message", () => {
    const wrapper = shallow(<AppList {...props} />);
    expect(wrapper.find(ErrorSelector)).toExist();
    expect(wrapper.find(ErrorSelector).html()).toContain("Sorry! Something went wrong.");
    expect(wrapper.find(ErrorSelector).html()).toContain(shallow(genericMessage).html());
    expect(wrapper).toMatchSnapshot();
  });

  it("renders a Application header", () => {
    const wrapper = shallow(<AppList {...props} />);
    expect(wrapper.find("h1").text()).toContain("Applications");
  });

  it("does not show the search filter nor show deleted checkbox nor deploy button", () => {
    const wrapper = shallow(<AppList {...props} />);
    expect(wrapper.find("SearchFilter")).not.toExist();
    expect(wrapper.find("label.checkbox")).not.toExist();
    expect(wrapper.find(".deploy-button")).not.toExist();
  });
});

context("when apps available", () => {
  beforeEach(() => {
    props = {
      ...defaultProps,
      apps: {
        listOverview: [
          {
            releaseName: "foo",
            chartMetadata: {
              name: "bar",
              version: "1.0.0",
              appVersion: "0.1.0",
            },
          } as IAppOverview,
        ],
      },
    };
  });

  it("matches the snapshot", () => {
    const wrapper = shallow(<AppList {...props} />);
    expect(wrapper).toMatchSnapshot();
  });

  it("renders a CardGrid with the available Apps", () => {
    const wrapper = shallow(<AppList {...props} />);
    const itemList = wrapper.find(AppListItem);
    expect(itemList).toExist();
    expect(itemList.key()).toBe("foo");
  });
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
              chartMetadata: {
                name: "foobar",
                version: "1.0.0",
                appVersion: "0.1.0",
              },
            } as IAppOverview,
            {
              releaseName: "bar",
              chartMetadata: {
                name: "foobar",
                version: "1.0.0",
                appVersion: "0.1.0",
              },
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
    listOverview: [
      {
        releaseName: "foo",
        chartMetadata: {
          name: "bar",
          version: "1.0.0",
          appVersion: "0.1.0",
        },
      } as IAppOverview,
    ],
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
  const fetchCalls = defaultProps.fetchAppsWithUpdateInfo.mock.calls;
  expect(fetchCalls[fetchCalls.length - 1]).toEqual(["default", true]);
});

it("renders the 'Show deleted apps' button even if the app list is empty", () => {
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
  expect(wrapper.find('input[type="checkbox"]').exists()).toBe(true);
});
