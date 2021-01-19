import * as React from "react";

import FilterGroup from "components/FilterGroup/FilterGroup";
import InfoCard from "components/InfoCard/InfoCard";
import Alert from "components/js/Alert";
import { act } from "react-dom/test-utils";
import { defaultStore, getStore, mountWrapper } from "shared/specs/mountWrapper";
import { IChart, IChartState, IClusterServiceVersion } from "../../shared/types";
import SearchFilter from "../SearchFilter/SearchFilter";
import Catalog, { filterNames } from "./Catalog";

const defaultChartState = {
  isFetching: false,
  selected: {} as IChartState["selected"],
  deployed: {} as IChartState["deployed"],
  items: [],
  searchItems: [],
  categories: [],
  updatesInfo: {},
} as IChartState;
const defaultProps = {
  charts: defaultChartState,
  repo: "",
  filter: {},
  fetchCharts: jest.fn(),
  fetchChartCategories: jest.fn(),
  pushSearchFilter: jest.fn(),
  cluster: "default",
  namespace: "kubeapps",
  kubeappsNamespace: "kubeapps",
  csvs: [],
  getCSVs: jest.fn(),
};
const chartItem = {
  id: "foo",
  attributes: {
    name: "foo",
    description: "",
    category: "",
    repo: { name: "foo", namespace: "chart-namespace" },
  },
  relationships: { latestChartVersion: { data: { app_version: "v1.0.0" } } },
} as IChart;
const chartItem2 = {
  id: "bar",
  attributes: {
    name: "bar",
    description: "",
    category: "Database",
    repo: { name: "bar", namespace: "chart-namespace" },
  },
  relationships: { latestChartVersion: { data: { app_version: "v2.0.0" } } },
} as IChart;
const csv = {
  metadata: {
    name: "test-csv",
  },
  spec: {
    provider: {
      name: "me",
    },
    icon: [{ base64data: "data", mediatype: "img/png" }],
    customresourcedefinitions: {
      owned: [
        {
          name: "foo-cluster",
          displayName: "foo-cluster",
          version: "v1.0.0",
          description: "a meaningful description",
        },
      ],
    },
  },
} as IClusterServiceVersion;
const populatedProps = {
  ...defaultProps,
  csvs: [csv],
  charts: { ...defaultChartState, items: [chartItem, chartItem2], searchItems: [chartItem] },
};

it("retrieves csvs in the namespace", () => {
  const getCSVs = jest.fn();
  mountWrapper(defaultStore, <Catalog {...populatedProps} getCSVs={getCSVs} />);
  expect(getCSVs).toHaveBeenCalledWith(defaultProps.cluster, defaultProps.namespace);
});

it("shows all the elements", () => {
  const wrapper = mountWrapper(defaultStore, <Catalog {...populatedProps} />);
  expect(wrapper.find(InfoCard)).toHaveLength(3);
});

it("should render a message if there are no elements in the catalog", () => {
  const wrapper = mountWrapper(defaultStore, <Catalog {...defaultProps} />);
  const message = wrapper.find(".empty-catalog");
  expect(message).toExist();
  expect(message).toIncludeText("The current catalog is empty");
});

it("should render an error if it exists", () => {
  const charts = {
    ...defaultChartState,
    selected: {
      error: new Error("Boom!"),
    },
  } as any;
  const wrapper = mountWrapper(defaultStore, <Catalog {...populatedProps} charts={charts} />);
  const error = wrapper.find(Alert);
  expect(error.prop("theme")).toBe("danger");
  expect(error).toIncludeText("Boom!");
});

it("behaves like a loading wrapper", () => {
  const wrapper = mountWrapper(
    defaultStore,
    <Catalog
      {...populatedProps}
      charts={{ isFetching: true, searchItems: [], items: [], categories: [], selected: {} } as any}
    />,
  );
  expect(wrapper.find("LoadingWrapper")).toExist();
});

describe("filters by the searched item", () => {
  it("modifying the search box invokes fetchCharts function", () => {
    const fetchCharts = jest.fn();
    const props = {
      ...populatedProps,
      fetchCharts,
    };
    const wrapper = mountWrapper(
      defaultStore,
      <Catalog {...props} filter={{ [filterNames.SEARCH]: "bar" }} />,
    );
    act(() => {
      (wrapper.find(SearchFilter).prop("onChange") as any)("bar");
      (wrapper.find(SearchFilter).prop("submitFilters") as any)("bar");
    });
    wrapper.update();
    expect(fetchCharts).toHaveBeenCalled();
  });
});

describe("filters by application type", () => {
  it("doesn't show the filter if there are no csvs", () => {
    const wrapper = mountWrapper(defaultStore, <Catalog {...populatedProps} csvs={[]} />);
    expect(
      wrapper.find(FilterGroup).findWhere(g => g.prop("name") === filterNames.TYPE),
    ).not.toExist();
  });

  it("filters only charts", () => {
    const wrapper = mountWrapper(
      defaultStore,
      <Catalog {...populatedProps} filter={{ Type: "Charts" }} />,
    );
    expect(wrapper.find(InfoCard)).toHaveLength(2);
  });

  it("push filter for only charts", () => {
    const store = getStore({});
    const wrapper = mountWrapper(store, <Catalog {...populatedProps} />);
    const input = wrapper.find("input").findWhere(i => i.prop("value") === "Charts");
    input.simulate("change", { target: { value: "Charts" } });
    // It should have pushed with the filter
    const historyAction = store
      .getActions()
      .find(action => action.type === "@@router/CALL_HISTORY_METHOD");
    expect(historyAction.payload).toEqual({
      args: ["/c/default/ns/kubeapps/catalog?Type=Charts"],
      method: "push",
    });
  });

  it("filters only operators", () => {
    const wrapper = mountWrapper(
      defaultStore,
      <Catalog {...populatedProps} filter={{ Type: "Operators" }} />,
    );
    expect(wrapper.find(InfoCard)).toHaveLength(1);
  });

  it("push filter for only operators", () => {
    const store = getStore({});
    const wrapper = mountWrapper(store, <Catalog {...populatedProps} />);
    const input = wrapper.find("input").findWhere(i => i.prop("value") === "Operators");
    input.simulate("change", { target: { value: "Operators" } });
    // It should have pushed with the filter
    const historyAction = store
      .getActions()
      .find(action => action.type === "@@router/CALL_HISTORY_METHOD");
    expect(historyAction.payload).toEqual({
      args: ["/c/default/ns/kubeapps/catalog?Type=Operators"],
      method: "push",
    });
  });
});

describe("filters by application repository", () => {
  it("doesn't show the filter if there are no apps", () => {
    const wrapper = mountWrapper(defaultStore, <Catalog {...defaultProps} />);
    expect(
      wrapper.find(FilterGroup).findWhere(g => g.prop("name") === filterNames.REPO),
    ).not.toExist();
  });

  it("filters by repo", () => {
    const wrapper = mountWrapper(
      defaultStore,
      <Catalog {...populatedProps} filter={{ [filterNames.REPO]: "foo" }} />,
    );
    expect(wrapper.find(InfoCard)).toHaveLength(1);
  });

  it("push filter for repo", () => {
    const store = getStore({});
    const wrapper = mountWrapper(store, <Catalog {...populatedProps} />);
    // The repo name is "foo"
    const input = wrapper.find("input").findWhere(i => i.prop("value") === "foo");
    input.simulate("change", { target: { value: "foo" } });
    // It should have pushed with the filter
    const historyAction = store
      .getActions()
      .find(action => action.type === "@@router/CALL_HISTORY_METHOD");
    expect(historyAction.payload).toEqual({
      args: ["/c/default/ns/kubeapps/catalog?Repository=foo"],
      method: "push",
    });
  });
});

describe("filters by operator provider", () => {
  it("doesn't show the filter if there are no csvs", () => {
    const wrapper = mountWrapper(defaultStore, <Catalog {...defaultProps} />);
    expect(
      wrapper.find(FilterGroup).findWhere(g => g.prop("name") === filterNames.OPERATOR_PROVIDER),
    ).not.toExist();
  });

  const csv2 = {
    metadata: {
      name: "csv2",
    },
    spec: {
      ...csv.spec,
      provider: {
        name: "you",
      },
    },
  } as any;

  it("push filter for operator provider", () => {
    const store = getStore({});
    const wrapper = mountWrapper(store, <Catalog {...populatedProps} csvs={[csv, csv2]} />);
    const input = wrapper.find("input").findWhere(i => i.prop("value") === "you");
    input.simulate("change", { target: { value: "you" } });
    // It should have pushed with the filter
    const historyAction = store
      .getActions()
      .find(action => action.type === "@@router/CALL_HISTORY_METHOD");
    expect(historyAction.payload).toEqual({
      args: ["/c/default/ns/kubeapps/catalog?Provider=you"],
      method: "push",
    });
  });

  it("filters by operator provider", () => {
    const wrapper = mountWrapper(
      defaultStore,
      <Catalog
        {...populatedProps}
        csvs={[csv, csv2]}
        filter={{ [filterNames.OPERATOR_PROVIDER]: "you" }}
      />,
    );
    expect(wrapper.find(InfoCard)).toHaveLength(1);
  });
});

describe("filters by category", () => {
  it("renders a Unknown category if not set", () => {
    const wrapper = mountWrapper(
      defaultStore,
      <Catalog
        {...defaultProps}
        charts={{
          ...defaultChartState,
          items: [chartItem],
          categories: [{ name: chartItem.attributes.category, count: 1 }],
        }}
      />,
    );
    expect(wrapper.find("input").findWhere(i => i.prop("value") === "Unknown")).toExist();
  });

  it("push filter for category", () => {
    const store = getStore({});
    const wrapper = mountWrapper(
      store,
      <Catalog
        {...defaultProps}
        charts={{
          ...defaultChartState,
          items: [chartItem, chartItem2],
          categories: [
            { name: chartItem.attributes.category, count: 1 },
            { name: chartItem2.attributes.category, count: 1 },
          ],
        }}
      />,
    );
    expect(wrapper.find(InfoCard)).toHaveLength(2);
    const input = wrapper.find("input").findWhere(i => i.prop("value") === "Database");
    input.simulate("change", { target: { value: "Database" } });
    // It should have pushed with the filter
    const historyAction = store
      .getActions()
      .find(action => action.type === "@@router/CALL_HISTORY_METHOD");
    expect(historyAction.payload).toEqual({
      args: ["/c/default/ns/kubeapps/catalog?Category=Database"],
      method: "push",
    });
  });

  it("filters a category", () => {
    const wrapper = mountWrapper(
      defaultStore,
      <Catalog
        {...defaultProps}
        charts={{
          ...defaultChartState,
          items: [chartItem, chartItem2],
          categories: [
            { name: chartItem.attributes.category, count: 1 },
            { name: chartItem2.attributes.category, count: 1 },
          ],
        }}
        filter={{ [filterNames.CATEGORY]: "Database" }}
      />,
    );
    expect(wrapper.find(InfoCard)).toHaveLength(1);
  });

  it("filters an operator category", () => {
    const csvWithCat = {
      ...csv,
      metadata: {
        name: "csv-cat",
        annotations: {
          categories: "E-Learning",
        },
      },
    } as any;
    const wrapper = mountWrapper(
      defaultStore,
      <Catalog
        {...defaultProps}
        csvs={[csv, csvWithCat]}
        filter={{ [filterNames.CATEGORY]: "E-Learning" }}
      />,
    );
    expect(wrapper.find(InfoCard)).toHaveLength(1);
  });

  it("filters operator categories", () => {
    const csvWithCat = {
      ...csv,
      metadata: {
        name: "csv-cat",
        annotations: {
          categories: "DeveloperTools, Infrastructure",
        },
      },
    } as any;
    const wrapper = mountWrapper(
      defaultStore,
      <Catalog
        {...defaultProps}
        csvs={[csv, csvWithCat]}
        filter={{ [filterNames.CATEGORY]: "Developer Tools,Infrastructure" }}
      />,
    );
    expect(wrapper.find(InfoCard)).toHaveLength(1);
  });
});
