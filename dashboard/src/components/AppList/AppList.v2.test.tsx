import { mount, shallow } from "enzyme";
import context from "jest-plugin-context";
import * as React from "react";

import SearchFilter from "components/SearchFilter/SearchFilter.v2";
import { BrowserRouter as Router } from "react-router-dom";
import itBehavesLike from "../../shared/specs";
import { IAppOverview, IAppState } from "../../shared/types";
import Alert from "../js/Alert";
import AppList, { IAppListProps } from "./AppList.v2";
import AppListItem from "./AppListItem.v2";
import CustomResourceListItem from "./CustomResourceListItem.v2";

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
  featureFlags: { operators: true, additionalClusters: [], ui: "hex" },
  appVersion: "v1.0.0",
};

context("when changing props", () => {
  it("should fetch apps in the new namespace", async () => {
    const fetchAppsWithUpdateInfo = jest.fn();
    const getCustomResources = jest.fn();
    mount(
      <Router>
        <AppList
          {...defaultProps}
          fetchAppsWithUpdateInfo={fetchAppsWithUpdateInfo}
          getCustomResources={getCustomResources}
          namespace="foo"
        />
      </Router>,
    );
    expect(fetchAppsWithUpdateInfo).toHaveBeenCalledWith("foo", true);
    expect(getCustomResources).toHaveBeenCalledWith("foo");
  });

  it("should update the filter", () => {
    const wrapper = mount(
      <Router>
        <AppList {...defaultProps} filter="foo" />
      </Router>,
    );
    expect(wrapper.find(SearchFilter).prop("value")).toEqual("foo");
  });
});

context("while fetching apps", () => {
  props = { ...defaultProps, apps: { isFetching: true } };

  itBehavesLike("aLoadingComponent", { component: AppList, props });
  itBehavesLike("aLoadingComponent", {
    component: AppList,
    props: { ...defaultProps, isFetchingResources: true },
  });

  it("renders a Application header", () => {
    const wrapper = shallow(<AppList {...props} />);
    expect(wrapper.find("h1").text()).toContain("Applications");
  });

  it("shows the search filter and deploy button", () => {
    const wrapper = shallow(<AppList {...props} />);
    expect(wrapper.find("SearchFilter")).toExist();
    expect(wrapper.find("Link").findWhere(l => l.text() === "Deploy")).toExist();
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
    const wrapper = mount(
      <Router>
        <AppList {...props} />
      </Router>,
    );
    expect(wrapper.find(".applist-empty").text()).toContain("Welcome To Kubeapps");
  });

  it("shows the search filter and deploy button", () => {
    const wrapper = shallow(<AppList {...props} />);
    expect(wrapper.find("SearchFilter")).toExist();
    expect(wrapper.find("Link").findWhere(l => l.text() === "Deploy")).toExist();
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
    expect(wrapper.find(Alert)).toExist();
    expect(wrapper.find(Alert).html()).toContain("Boom!");
    expect(wrapper).toMatchSnapshot();
  });

  it("renders a Application header", () => {
    const wrapper = shallow(<AppList {...props} />);
    expect(wrapper.find("h1").text()).toContain("Applications");
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
            status: "deployed",
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
    const wrapper = mount(
      <Router>
        <AppList {...props} />
      </Router>,
    );
    const itemList = wrapper.find(AppListItem);
    expect(itemList).toExist();
    expect(itemList.key()).toBe("foo");
  });
});

it("filters apps", () => {
  const wrapper = mount(
    <Router>
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
                status: "deployed",
              } as IAppOverview,
              {
                releaseName: "bar",
                chartMetadata: {
                  name: "foobar",
                  version: "1.0.0",
                  appVersion: "0.1.0",
                },
                status: "deployed",
              } as IAppOverview,
            ],
            listingAll: false,
          } as IAppState
        }
        filter="bar"
      />
      ,
    </Router>,
  );
  expect(wrapper.find(AppListItem).key()).toBe("bar");
});

context("when custom resources available", () => {
  beforeEach(() => {
    const cr = { kind: "KubeappsCluster", metadata: { name: "foo-cluster" } };
    const csv = {
      metadata: {
        name: "foo",
      },
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
    const wrapper = mount(
      <Router>
        <AppList {...props} />
      </Router>,
    );
    const itemList = wrapper.find(CustomResourceListItem);
    expect(itemList).toExist();
    expect(itemList.key()).toBe("foo-cluster");
  });

  it("filters out items", () => {
    const wrapper = mount(
      <Router>
        <AppList {...props} filter="nop" />
      </Router>,
    );
    const itemList = wrapper.find(CustomResourceListItem);
    expect(itemList).not.toExist();
  });
});
