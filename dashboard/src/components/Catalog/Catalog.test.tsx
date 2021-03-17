import FilterGroup from "components/FilterGroup/FilterGroup";
import InfoCard from "components/InfoCard/InfoCard";
import Alert from "components/js/Alert";
import LoadingWrapper from "components/LoadingWrapper";
import React from "react";
import { act } from "react-dom/test-utils";
import * as ReactRedux from "react-redux";
import { defaultStore, getStore, initialState, mountWrapper } from "shared/specs/mountWrapper";
import { IAppRepository, IChart, IChartState, IClusterServiceVersion } from "../../shared/types";
import SearchFilter from "../SearchFilter/SearchFilter";
import Catalog, { filterNames } from "./Catalog";
import CatalogItems from "./CatalogItems";
import ChartCatalogItem from "./ChartCatalogItem";

const defaultChartState = {
  isFetching: false,
  hasFinishedFetching: false,
  selected: {} as IChartState["selected"],
  deployed: {} as IChartState["deployed"],
  items: [],
  categories: [],
  size: 20,
} as IChartState;
const defaultProps = {
  charts: defaultChartState,
  repo: "",
  filter: {},
  fetchCharts: jest.fn(),
  fetchChartCategories: jest.fn(),
  fetchRepos: jest.fn(),
  resetRequestCharts: jest.fn(),
  pushSearchFilter: jest.fn(),
  cluster: initialState.config.kubeappsCluster,
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
const populatedChartProps = { ...defaultChartState, items: [chartItem, chartItem2] };
const populatedProps = {
  ...defaultProps,
  csvs: [csv],
  charts: populatedChartProps,
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

it("should not render a message if there are no elements in the catalog but the fetching hasn't ended", () => {
  const wrapper = mountWrapper(defaultStore, <Catalog {...defaultProps} />);
  const message = wrapper.find(".empty-catalog");
  expect(message).not.toExist();
  expect(message).not.toIncludeText("The current catalog is empty");
});

it("should render a message if there are no elements in the catalog and the fetching has ended", () => {
  const wrapper = mountWrapper(
    defaultStore,
    <Catalog {...defaultProps} charts={{ ...defaultChartState, hasFinishedFetching: true }} />,
  );
  wrapper.setProps({ searchFilter: "" });
  const message = wrapper.find(".empty-catalog");
  expect(message).toExist();
  expect(message).toIncludeText("The current catalog is empty");
});

it("should render a spinner if there are no elements but it's still fetching", () => {
  const wrapper = mountWrapper(
    defaultStore,
    <Catalog {...defaultProps} charts={{ ...defaultChartState, hasFinishedFetching: false }} />,
  );
  expect(wrapper.find(LoadingWrapper)).toExist();
});

it("should not render a spinner if there are no elements and it finished fetching", () => {
  const wrapper = mountWrapper(
    defaultStore,
    <Catalog {...defaultProps} charts={{ ...defaultChartState, hasFinishedFetching: true }} />,
  );
  expect(wrapper.find(LoadingWrapper)).not.toExist();
});

it("should render a spinner if there already pending elements", () => {
  const wrapper = mountWrapper(
    defaultStore,
    <Catalog {...populatedProps} charts={{ ...populatedChartProps, hasFinishedFetching: false }} />,
  );
  expect(wrapper.find(LoadingWrapper)).toExist();
});

it("should not render a message if only operators are selected", () => {
  const wrapper = mountWrapper(
    defaultStore,
    <Catalog
      {...populatedProps}
      charts={{ ...populatedChartProps, hasFinishedFetching: true }}
      filter={{ [filterNames.TYPE]: "Operators" }}
    />,
  );
  expect(wrapper.find(LoadingWrapper)).not.toExist();
});

it("should not render a message if there are no more elements", () => {
  const wrapper = mountWrapper(
    defaultStore,
    <Catalog {...populatedProps} charts={{ ...populatedChartProps, hasFinishedFetching: true }} />,
  );
  const message = wrapper.find(".endPageMessage");
  expect(message).not.toExist();
});

it("should not render a message if there are no more elements but it's searching", () => {
  const wrapper = mountWrapper(
    defaultStore,
    <Catalog
      {...populatedProps}
      charts={{ ...populatedChartProps, hasFinishedFetching: true }}
      filter={{ [filterNames.SEARCH]: "bar" }}
    />,
  );
  const message = wrapper.find(".endPageMessage");
  expect(message).not.toExist();
});

it("should render the scroll handler if not finished", () => {
  const wrapper = mountWrapper(
    defaultStore,
    <Catalog {...populatedProps} charts={{ ...populatedChartProps, hasFinishedFetching: false }} />,
  );
  const scroll = wrapper.find(".scrollHandler");
  expect(scroll).toExist();
  expect(scroll).toHaveProperty("ref");
});

it("should not render the scroll handler if finished", () => {
  const wrapper = mountWrapper(
    defaultStore,
    <Catalog {...populatedProps} charts={{ ...populatedChartProps, hasFinishedFetching: true }} />,
  );
  const scroll = wrapper.find(".scrollHandler");
  expect(scroll).not.toExist();
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
      charts={{ isFetching: true, items: [], categories: [], selected: {} } as any}
    />,
  );
  expect(wrapper.find(LoadingWrapper)).toExist();
});

it("transforms the received '__' in query params into a ','", () => {
  const wrapper = mountWrapper(
    defaultStore,
    <Catalog
      {...populatedProps}
      filter={{ [filterNames.OPERATOR_PROVIDER]: "Lightbend__%20Inc." }}
    />,
  );
  expect(wrapper.find(".label-info").text()).toBe("Provider: Lightbend,%20Inc. ");
});

describe("filters by the searched item", () => {
  let spyOnUseDispatch: jest.SpyInstance;
  let spyOnUseEffect: jest.SpyInstance;

  afterEach(() => {
    spyOnUseDispatch.mockRestore();
    spyOnUseEffect.mockRestore();
  });

  it("filters modifying the search box", () => {
    const fetchCharts = jest.fn();
    const resetRequestCharts = jest.fn();
    const mockDispatch = jest.fn();
    const mockUseEffect = jest.fn();

    spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
    spyOnUseEffect = jest.spyOn(React, "useEffect").mockReturnValue(mockUseEffect as any);

    const props = {
      ...populatedProps,
      fetchCharts,
      resetRequestCharts,
    };
    const wrapper = mountWrapper(
      defaultStore,
      <Catalog {...props} filter={{ [filterNames.SEARCH]: "bar" }} />,
    );
    act(() => {
      (wrapper.find(SearchFilter).prop("onChange") as any)("bar");
    });
    wrapper.update();
    expect(mockDispatch).toHaveBeenCalledWith({
      payload: {
        args: ["/c/default-cluster/ns/kubeapps/catalog?Search=bar"],
        method: "push",
      },
      type: "@@router/CALL_HISTORY_METHOD",
    });
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
      args: ["/c/default-cluster/ns/kubeapps/catalog?Type=Charts"],
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
      args: ["/c/default-cluster/ns/kubeapps/catalog?Type=Operators"],
      method: "push",
    });
  });
});

describe("pagination and chart fetching", () => {
  let spyOnUseState: jest.SpyInstance;

  it("sets the initial state page to 1 before fetching charts", () => {
    const fetchCharts = jest.fn();
    const resetRequestCharts = jest.fn();

    const wrapper = mountWrapper(
      defaultStore,
      <Catalog
        {...populatedProps}
        fetchCharts={fetchCharts}
        resetRequestCharts={resetRequestCharts}
        charts={
          {
            ...defaultChartState,
            hasFinishedFetching: false,
            isFetching: false,
            items: [],
          } as any
        }
      />,
    );

    expect(wrapper.find(CatalogItems).prop("page")).toBe(1);
    expect(wrapper.find(ChartCatalogItem).length).toBe(0);
    expect(fetchCharts).toHaveBeenNthCalledWith(1, "default-cluster", "kubeapps", "", 1, 20, "");
    expect(resetRequestCharts).toHaveBeenNthCalledWith(1);
  });

  it("sets the state page when fetching charts", () => {
    const fetchCharts = jest.fn();
    const resetRequestCharts = jest.fn();

    const wrapper = mountWrapper(
      defaultStore,
      <Catalog
        {...populatedProps}
        fetchCharts={fetchCharts}
        resetRequestCharts={resetRequestCharts}
        charts={
          {
            ...defaultChartState,
            hasFinishedFetching: false,
            isFetching: true,
            items: [chartItem],
          } as any
        }
      />,
    );
    expect(wrapper.find(CatalogItems).prop("page")).toBe(1);
    expect(wrapper.find(ChartCatalogItem).length).toBe(0);
    expect(fetchCharts).toHaveBeenCalledWith("default-cluster", "kubeapps", "", 1, 20, "");
    expect(resetRequestCharts).toHaveBeenCalledWith();
  });

  it("items are translated to CatalogItems after fetching charts", () => {
    const fetchCharts = jest.fn();
    const resetRequestCharts = jest.fn();

    const wrapper = mountWrapper(
      defaultStore,
      <Catalog
        {...populatedProps}
        fetchCharts={fetchCharts}
        resetRequestCharts={resetRequestCharts}
        charts={
          {
            ...defaultChartState,
            hasFinishedFetching: true,
            isFetching: false,
            items: [chartItem, chartItem2],
          } as any
        }
      />,
    );
    expect(wrapper.find(CatalogItems).prop("page")).toBe(1);
    expect(wrapper.find(ChartCatalogItem).length).toBe(2);
    expect(fetchCharts).toHaveBeenCalledWith("default-cluster", "kubeapps", "", 1, 20, "");
    expect(resetRequestCharts).toHaveBeenCalledWith();
  });

  it("changes page", () => {
    const setState = jest.fn();
    const setPage = jest.fn();
    const charts = {
      ...defaultChartState,
      hasFinishedFetching: false,
      isFetching: false,
      items: [],
    } as any;
    spyOnUseState = jest
      .spyOn(React, "useState")
      /* @ts-expect-error: Argument of type '(init: any) => any' is not assignable to parameter of type '() => [unknown, Dispatch<unknown>]' */
      .mockImplementation((init: any) => {
        if (init === false) {
          // Mocking the result of hasLoadedFirstPage to simulate that is already loaded
          return [true, setState];
        }
        if (init === 1) {
          // Mocking the result of setPage to ensure it's called
          return [1, setPage];
        }
        return [init, setState];
      });

    // Mock intersection observer
    const observe = jest.fn();
    const unobserve = jest.fn();

    window.IntersectionObserver = jest.fn(callback => {
      (callback as (e: any) => void)([{ isIntersecting: true }]);
      return { observe, unobserve } as any;
    });
    window.IntersectionObserverEntry = jest.fn();

    mountWrapper(defaultStore, <Catalog {...populatedProps} charts={charts} />);
    spyOnUseState.mockRestore();
    expect(setPage).toHaveBeenCalledWith(2);
  });

  // TODO(agamez): add a test case covering it "resets page when one of the filters changes"
  // https://github.com/kubeapps/kubeapps/pull/2264/files/0d3c77448543668255809bf05039aca704cf729f..22343137efb1c2292b0aa4795f02124306cb055e#r565486271
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
    const store = getStore({ repos: { repos: [{ metadata: { name: "foo" } } as IAppRepository] } });
    const fetchRepos = jest.fn();
    const wrapper = mountWrapper(store, <Catalog {...populatedProps} fetchRepos={fetchRepos} />);
    // The repo name is "foo"
    const input = wrapper.find("input").findWhere(i => i.prop("value") === "foo");
    input.simulate("change", { target: { value: "foo" } });
    // It should have pushed with the filter
    const historyAction = store
      .getActions()
      .find(action => action.type === "@@router/CALL_HISTORY_METHOD");
    expect(fetchRepos).toHaveBeenCalledWith("kubeapps");
    expect(historyAction.payload).toEqual({
      args: ["/c/default-cluster/ns/kubeapps/catalog?Repository=foo"],
      method: "push",
    });
  });
});

it("push filter for repo", () => {
  const store = getStore({
    ...defaultStore,
    repos: { repos: [{ metadata: { name: "foo" } } as IAppRepository] },
  });
  const fetchRepos = jest.fn();
  const wrapper = mountWrapper(store, <Catalog {...populatedProps} fetchRepos={fetchRepos} />);
  // The repo name is "foo"
  const input = wrapper.find("input").findWhere(i => i.prop("value") === "foo");
  input.simulate("change", { target: { value: "foo" } });
  // It should have pushed with the filter
  const historyAction = store
    .getActions()
    .find(action => action.type === "@@router/CALL_HISTORY_METHOD");
  expect(fetchRepos).toHaveBeenCalledWith("kubeapps");
  expect(historyAction.payload).toEqual({
    args: ["/c/default-cluster/ns/kubeapps/catalog?Repository=foo"],
    method: "push",
  });
});

it("push filter for repo in other ns", () => {
  const store = getStore({
    ...defaultStore,
    repos: { repos: [{ metadata: { name: "foo" } } as IAppRepository] },
  });
  const fetchRepos = jest.fn();
  const wrapper = mountWrapper(
    store,
    <Catalog {...populatedProps} namespace={"my-ns"} fetchRepos={fetchRepos} />,
  );
  // The repo name is "foo", the ns name is "my-ns"
  const input = wrapper.find("input").findWhere(i => i.prop("value") === "foo");
  input.simulate("change", { target: { value: "foo" } });
  // It should have pushed with the filter
  const historyAction = store
    .getActions()
    .find(action => action.type === "@@router/CALL_HISTORY_METHOD");
  expect(fetchRepos).toHaveBeenCalledWith("my-ns", true);
  expect(historyAction.payload).toEqual({
    args: ["/c/default-cluster/ns/my-ns/catalog?Repository=foo"],
    method: "push",
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
      args: ["/c/default-cluster/ns/kubeapps/catalog?Provider=you"],
      method: "push",
    });
  });

  it("push filter for operator provider with comma", () => {
    const store = getStore({});
    const wrapper = mountWrapper(store, <Catalog {...populatedProps} csvs={[csv, csv2]} />);
    const input = wrapper.find("input").findWhere(i => i.prop("value") === "you");
    input.simulate("change", { target: { value: "you, inc" } });
    // It should have pushed with the filter
    const historyAction = store
      .getActions()
      .find(action => action.type === "@@router/CALL_HISTORY_METHOD");
    expect(historyAction.payload).toEqual({
      args: ["/c/default-cluster/ns/kubeapps/catalog?Provider=you__%20inc"],
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
      args: ["/c/default-cluster/ns/kubeapps/catalog?Category=Database"],
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
