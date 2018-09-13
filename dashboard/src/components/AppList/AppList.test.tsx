import { shallow } from "enzyme";
import * as React from "react";

import sharedSpecs from "../../shared/specs";
import { IAppOverview, IAppState } from "../../shared/types";
import { CardGrid } from "../Card";
import AppList from "./AppList";
import AppListItem from "./AppListItem";

let defaultProps = {} as any;
let props = {} as any;

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

describe("while fetching apps", () => {
  beforeEach(() => {
    props = { ...defaultProps, apps: { isFetching: true } };
  });

  afterAll(() => {
    // Calling the shared specs here so they are run after all the beforeEach
    // callbacks have been run which set the default props and prop
    sharedSpecs.loadingSpecs(AppList, props);
  });

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

describe("when fetched but not apps available", () => {
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

describe("when error present", () => {
  beforeEach(() => {
    props = { ...defaultProps, apps: { error: "Boom!" } };
  });

  it("matches the snapshot", () => {
    const wrapper = shallow(<AppList {...props} />);
    expect(wrapper).toMatchSnapshot();
  });

  it("renders a generic error message", () => {
    const wrapper = shallow(<AppList {...props} />);
    expect(wrapper.find("UnexpectedErrorPage")).toExist();
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

describe("when apps available", () => {
  beforeEach(() => {
    props = {
      ...defaultProps,
      apps: {
        listOverview: [
          {
            releaseName: "foo",
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
