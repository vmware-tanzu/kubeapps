import { defaultStore, mountWrapper } from "shared/specs/mountWrapper";
import { IChart, IClusterServiceVersion } from "shared/types";
import CatalogItem from "./CatalogItem";
import CatalogItems from "./CatalogItems";

const chartItem = {
  id: "foo",
  attributes: {
    name: "foo",
    description: "",
    repo: { name: "foo", namespace: "chart-namespace" },
  },
  relationships: { latestChartVersion: { data: { app_version: "v1.0.0" } } },
} as IChart;
const chartItem2 = {
  id: "bar",
  attributes: {
    name: "bar",
    description: "",
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
const defaultProps = {
  charts: [],
  csvs: [],
  cluster: "default",
  namespace: "default",
  isFetching: false,
  page: 1,
  hasFinishedFetching: true,
};
const populatedProps = {
  ...defaultProps,
  charts: [chartItem, chartItem2],
  csvs: [csv],
};

it("shows nothing if no items are passed but it's still fetching", () => {
  const wrapper = mountWrapper(defaultStore, <CatalogItems {...defaultProps} isFetching={true} />);
  expect(wrapper).toIncludeText("");
});

it("shows a message if no items are passed and it stopped fetching", () => {
  const wrapper = mountWrapper(defaultStore, <CatalogItems {...defaultProps} isFetching={false} />);
  expect(wrapper).toIncludeText("No application matches the current filter");
});

it("no items if it's fetching and it's the first page (prevents showing incomplete list during the first render)", () => {
  const wrapper = mountWrapper(
    defaultStore,
    <CatalogItems {...populatedProps} isFetching={true} page={1} />,
  );
  const items = wrapper.find(CatalogItem);
  expect(items).toHaveLength(0);
});

it("show items if it's fetching but it is NOT the first page (allow pagination without scrolling issues)", () => {
  const wrapper = mountWrapper(
    defaultStore,
    <CatalogItems {...populatedProps} isFetching={true} page={2} />,
  );
  const items = wrapper.find(CatalogItem);
  expect(items).toHaveLength(3);
});

it("order elements by name", () => {
  const wrapper = mountWrapper(defaultStore, <CatalogItems {...populatedProps} />);
  const items = wrapper.find(CatalogItem).map(i => i.prop("item").name);
  expect(items).toEqual(["bar", "foo", "foo-cluster"]);
});
