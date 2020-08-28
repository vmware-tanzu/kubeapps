import context from "jest-plugin-context";
import * as React from "react";

import FilterGroup from "components/FilterGroup/FilterGroup";
import InfoCard from "components/InfoCard/InfoCard.v2";
import Alert from "components/js/Alert";
import { act } from "react-dom/test-utils";
import { defaultStore, mountWrapper } from "shared/specs/mountWrapper";
import itBehavesLike from "../../shared/specs";
import { IChart, IChartState, IClusterServiceVersion } from "../../shared/types";
import SearchFilter from "../SearchFilter/SearchFilter.v2";
import Catalog from "./Catalog.v2";

const defaultChartState = {
  isFetching: false,
  selected: {} as IChartState["selected"],
  deployed: {} as IChartState["deployed"],
  items: [],
  updatesInfo: {},
} as IChartState;
const defaultProps = {
  charts: defaultChartState,
  repo: "",
  filter: "",
  fetchCharts: jest.fn(),
  pushSearchFilter: jest.fn(),
  cluster: "default",
  namespace: "kubeapps",
  kubeappsNamespace: "kubeapps",
  csvs: [],
  getCSVs: jest.fn(),
  featureFlags: { operators: false, ui: "hex" },
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
  charts: { ...defaultChartState, items: [chartItem, chartItem2] },
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

context("when fetching apps", () => {
  itBehavesLike("aLoadingComponent", {
    component: Catalog,
    props: { ...defaultProps, charts: { isFetching: true, items: [], selected: {} } },
  });
});

describe("filters by the searched item", () => {
  it("filters using prop", () => {
    const wrapper = mountWrapper(defaultStore, <Catalog {...populatedProps} filter={"bar"} />);
    expect(wrapper.find(InfoCard)).toHaveLength(1);
  });

  it("filters modifying the search box", () => {
    const wrapper = mountWrapper(defaultStore, <Catalog {...populatedProps} />);
    act(() => {
      (wrapper.find(SearchFilter).prop("onChange") as any)("bar");
    });
    wrapper.update();
    expect(wrapper.find(InfoCard)).toHaveLength(1);
  });
});

describe("filters by application type", () => {
  it("doesn't show the filter if there are no csvs", () => {
    const wrapper = mountWrapper(defaultStore, <Catalog {...populatedProps} csvs={[]} />);
    expect(wrapper.find(FilterGroup).findWhere(g => g.prop("name") === "apptype")).not.toExist();
  });

  it("filters only charts", () => {
    const wrapper = mountWrapper(defaultStore, <Catalog {...populatedProps} />);
    const input = wrapper.find("input").findWhere(i => i.prop("value") === "Charts");
    input.simulate("change", { target: { value: "Charts" } });
    wrapper.update();
    expect(wrapper.find(InfoCard)).toHaveLength(2);
  });

  it("filters only operators", () => {
    const wrapper = mountWrapper(defaultStore, <Catalog {...populatedProps} />);
    const input = wrapper.find("input").findWhere(i => i.prop("value") === "Operators");
    input.simulate("change", { target: { value: "Operators" } });
    wrapper.update();
    expect(wrapper.find(InfoCard)).toHaveLength(1);
  });
});

describe("filters by application repository", () => {
  it("doesn't show the filter if there are no apps", () => {
    const wrapper = mountWrapper(defaultStore, <Catalog {...defaultProps} />);
    expect(wrapper.find(FilterGroup).findWhere(g => g.prop("name") === "apprepo")).not.toExist();
  });

  it("filters by repo", () => {
    const wrapper = mountWrapper(defaultStore, <Catalog {...populatedProps} />);
    // The repo name is "foo"
    const input = wrapper.find("input").findWhere(i => i.prop("value") === "foo");
    input.simulate("change", { target: { value: "foo" } });
    wrapper.update();
    expect(wrapper.find(InfoCard)).toHaveLength(1);
  });
});

describe("filters by operator provider", () => {
  it("doesn't show the filter if there are no csvs", () => {
    const wrapper = mountWrapper(defaultStore, <Catalog {...defaultProps} />);
    expect(
      wrapper.find(FilterGroup).findWhere(g => g.prop("name") === "operator-provider"),
    ).not.toExist();
  });

  it("filters by operator provider", () => {
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
    const wrapper = mountWrapper(defaultStore, <Catalog {...populatedProps} csvs={[csv, csv2]} />);
    // The repo name is "foo"
    const input = wrapper.find("input").findWhere(i => i.prop("value") === "you");
    input.simulate("change", { target: { value: "you" } });
    wrapper.update();
    expect(wrapper.find(InfoCard)).toHaveLength(1);
  });
});

describe("filters by category", () => {
  it("renders a Unknown category if not set", () => {
    const wrapper = mountWrapper(
      defaultStore,
      <Catalog {...defaultProps} charts={{ ...defaultChartState, items: [chartItem] }} />,
    );
    expect(wrapper.find("input").findWhere(i => i.prop("value") === "Unknown")).toExist();
  });

  it("filters a category", () => {
    const wrapper = mountWrapper(
      defaultStore,
      <Catalog
        {...defaultProps}
        charts={{ ...defaultChartState, items: [chartItem, chartItem2] }}
      />,
    );
    expect(wrapper.find(InfoCard)).toHaveLength(2);
    const input = wrapper.find("input").findWhere(i => i.prop("value") === "Database");
    input.simulate("change", { target: { value: "Database" } });
    wrapper.update();
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
      <Catalog {...defaultProps} csvs={[csv, csvWithCat]} />,
    );
    expect(wrapper.find(InfoCard)).toHaveLength(2);

    const input = wrapper.find("input").findWhere(i => i.prop("value") === "E-Learning");
    input.simulate("change", { target: { value: "E-Learning" } });
    wrapper.update();
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
      <Catalog {...defaultProps} csvs={[csv, csvWithCat]} />,
    );
    expect(wrapper.find(InfoCard)).toHaveLength(2);

    // Two categories extracted from the same CSV
    expect(wrapper.find("input").findWhere(i => i.prop("value") === "Developer Tools")).toExist();
    const input = wrapper.find("input").findWhere(i => i.prop("value") === "Infrastructure");
    input.simulate("change", { target: { value: "Infrastructure" } });
    wrapper.update();
    expect(wrapper.find(InfoCard)).toHaveLength(1);
  });
});
