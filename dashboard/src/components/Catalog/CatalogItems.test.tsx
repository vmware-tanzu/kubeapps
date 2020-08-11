import * as React from "react";

import { defaultStore, mountWrapper } from "shared/specs/mountWrapper";
import { IChart, IClusterServiceVersion } from "shared/types";
import CatalogItem from "./CatalogItem.v2";
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
};
const populatedProps = {
  ...defaultProps,
  charts: [chartItem, chartItem2],
  csvs: [csv],
};

it("shows a message if no items are passed", () => {
  const wrapper = mountWrapper(defaultStore, <CatalogItems {...defaultProps} />);
  expect(wrapper).toIncludeText("No application matches the current filter");
});

it("order elements by name", () => {
  const wrapper = mountWrapper(defaultStore, <CatalogItems {...populatedProps} />);
  const items = wrapper.find(CatalogItem).map(i => i.prop("item").name);
  expect(items).toEqual(["bar", "foo", "foo-cluster"]);
});
