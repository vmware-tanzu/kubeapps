import { shallow } from "enzyme";
import context from "jest-plugin-context";
import * as React from "react";

import itBehavesLike from "../../shared/specs";
import { IAppOverview, IAppState } from "../../shared/types";
import { CardGrid } from "../Card";
import { ErrorSelector } from "../ErrorAlert";
import { genericMessage } from "../ErrorAlert/UnexpectedErrorAlert";
import AppList, { IAppListProps } from "./AppList";
import AppListItem from "./AppListItem";
import CustomResourceListItem from "./CustomResourceListItem";

let props = {} as any;

const defaultProps: IAppListProps = {
  apps: {} as IAppState,
  filter: "",
  namespace: "default",
  cluster: "defaultc",
  pushSearchFilter: jest.fn(),
  fetchAppsWithUpdateInfo: jest.fn(),
  getCustomResources: jest.fn(),
  isFetchingResources: false,
  customResources: [],
  csvs: [],
  featureFlags: { operators: true, ui: "hex" },
};

context("when changing props", () => {
  it("should fetch apps in the new namespace", () => {
    const fetchAppsWithUpdateInfo = jest.fn();
    const getCustomResources = jest.fn();
    const wrapper = shallow(
      <AppList
        {...defaultProps}
        fetchAppsWithUpdateInfo={fetchAppsWithUpdateInfo}
        getCustomResources={getCustomResources}
      />,
    );
    wrapper.setProps({ namespace: "foo" });
    expect(fetchAppsWithUpdateInfo).toHaveBeenCalledWith("defaultc", "foo", undefined);
    expect(getCustomResources).toHaveBeenCalledWith("foo");
  });

  it("should update the filter", () => {
    const wrapper = shallow(<AppList {...defaultProps} />);
    wrapper.setProps({ filter: "foo" });
    expect(wrapper.state("filter")).toEqual("foo");
  });
});

context("while fetching apps", () => {
  props = { ...defaultProps, apps: { isFetching: true } };

  itBehavesLike("aLoadingComponent", { component: AppList, props });
  itBehavesLike("aLoadingComponent", {
    component: AppList,
    props: { ...defaultProps, isFetchingResources: true },
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
  const mockFetchAppsWithUpdateInfo = jest.fn();
  const updatedProps: IAppListProps = {
    ...defaultProps,
    fetchAppsWithUpdateInfo: mockFetchAppsWithUpdateInfo,
  };
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
  const wrapper = shallow(<AppList {...updatedProps} apps={apps} />);
  const checkbox = wrapper.find('input[type="checkbox"]');
  expect(apps.listingAll).toBe(false);
  checkbox.simulate("change");
  // The last call to fetchApps should list all the apps
  const fetchCalls = mockFetchAppsWithUpdateInfo.mock.calls;
  expect(fetchCalls[fetchCalls.length - 1]).toEqual(["defaultc", "default", true]);
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

context("when custom resources available", () => {
  beforeEach(() => {
    const cr = { kind: "KubeappsCluster", metadata: { name: "foo-cluster" } };
    const csv = {
      spec: {
        customresourcedefinitions: {
          owned: [
            {
              kind: "KubeappsCluster",
            },
          ],
        },
      },
    };
    props = {
      ...defaultProps,
      customResources: [cr],
      csvs: [csv],
    };
  });

  it("matches the snapshot", () => {
    const wrapper = shallow(<AppList {...props} />);
    expect(wrapper).toMatchSnapshot();
  });

  it("renders a CardGrid with the available resources", () => {
    const wrapper = shallow(<AppList {...props} />);
    const itemList = wrapper.find(CustomResourceListItem);
    expect(itemList).toExist();
    expect(itemList.key()).toBe("foo-cluster");
  });

  it("filters out items", () => {
    const wrapper = shallow(<AppList {...props} />);
    wrapper.setState({ filter: "nop" });
    const itemList = wrapper.find(CustomResourceListItem);
    expect(itemList).not.toExist();
  });
});
