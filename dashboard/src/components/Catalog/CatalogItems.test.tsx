// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import InfoCard from "components/InfoCard";
import { AvailablePackageSummary, Context } from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import { defaultStore, mountWrapper } from "shared/specs/mountWrapper";
import { IClusterServiceVersion, PluginNames } from "shared/types";
import CatalogItem from "./CatalogItem";
import CatalogItems, { ICatalogItemsProps } from "./CatalogItems";

const availablePackageSummary1: AvailablePackageSummary = {
  name: "foo",
  categories: [],
  displayName: "foo",
  iconUrl: "",
  latestVersion: { appVersion: "v1.0.0", pkgVersion: "" },
  shortDescription: "",
  availablePackageRef: {
    identifier: "foo/foo",
    context: { cluster: "", namespace: "package-namespace" } as Context,
    plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
  },
};
const availablePackageSummary2: AvailablePackageSummary = {
  name: "bar",
  categories: ["Database"],
  displayName: "bar",
  iconUrl: "",
  latestVersion: { appVersion: "v2.0.0", pkgVersion: "" },
  shortDescription: "",
  availablePackageRef: {
    identifier: "bar/bar",
    context: { cluster: "", namespace: "package-namespace" } as Context,
    plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
  },
};
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
  availablePackageSummaries: [],
  csvs: [],
  cluster: "default",
  namespace: "default",
  hasLoadedFirstPage: true,
  isFirstPage: true,
  hasFinishedFetching: true,
} as ICatalogItemsProps;
const populatedProps = {
  ...defaultProps,
  availablePackageSummaries: [availablePackageSummary1, availablePackageSummary2],
  csvs: [csv],
} as ICatalogItemsProps;

it("shows nothing if no items are passed but it's still fetching", () => {
  const wrapper = mountWrapper(
    defaultStore,
    <CatalogItems {...defaultProps} hasLoadedFirstPage={false} />,
  );
  expect(wrapper).toIncludeText("");
});

it("shows a message if no items are passed and it stopped fetching", () => {
  const wrapper = mountWrapper(
    defaultStore,
    <CatalogItems {...defaultProps} hasLoadedFirstPage={true} />,
  );
  expect(wrapper).toIncludeText("No application matches the current filter");
});

it("no items if it's fetching and it's the first page (prevents showing incomplete list during the first render)", () => {
  const wrapper = mountWrapper(
    defaultStore,
    <CatalogItems {...populatedProps} hasLoadedFirstPage={false} isFirstPage={true} />,
  );
  const items = wrapper.find(CatalogItem);
  expect(items).toHaveLength(0);
});

it("show items if it's fetching but it is NOT the first page (allow pagination without scrolling issues)", () => {
  const wrapper = mountWrapper(
    defaultStore,
    <CatalogItems {...populatedProps} hasLoadedFirstPage={false} isFirstPage={false} />,
  );
  const items = wrapper.find(CatalogItem);
  expect(items).toHaveLength(3);
});

it("order elements by name", () => {
  const wrapper = mountWrapper(defaultStore, <CatalogItems {...populatedProps} />);
  const items = wrapper.find(CatalogItem).map(i => i.prop("item").name);
  expect(items).toEqual(["bar", "foo", "foo-cluster"]);
});

it("changes the bgIcon based on the plugin name - default", () => {
  const pluginName = "my.plugin";
  const populatedProps = {
    ...defaultProps,
    availablePackageSummaries: [
      {
        ...availablePackageSummary1,
        availablePackageRef: {
          ...availablePackageSummary1.availablePackageRef,
          plugin: { ...availablePackageSummary1.availablePackageRef?.plugin, name: pluginName },
        },
      } as AvailablePackageSummary,
    ],
  } as ICatalogItemsProps;

  const wrapper = mountWrapper(defaultStore, <CatalogItems {...populatedProps} />);
  expect(
    wrapper
      .find(InfoCard)
      .findWhere(s => s.prop("link")?.includes(pluginName))
      .prop("bgIcon"),
  ).toBe("placeholder.svg");
});

it("changes the bgIcon based on the plugin name - helm", () => {
  const pluginName = PluginNames.PACKAGES_HELM;
  const populatedProps = {
    ...defaultProps,
    availablePackageSummaries: [
      {
        ...availablePackageSummary1,
        availablePackageRef: {
          ...availablePackageSummary1.availablePackageRef,
          plugin: { ...availablePackageSummary1.availablePackageRef?.plugin, name: pluginName },
        },
      } as AvailablePackageSummary,
    ],
  } as ICatalogItemsProps;

  const wrapper = mountWrapper(defaultStore, <CatalogItems {...populatedProps} />);
  expect(
    wrapper
      .find(InfoCard)
      .findWhere(s => s.prop("link")?.includes(pluginName))
      .prop("bgIcon"),
  ).toBe("helm.svg");
});
