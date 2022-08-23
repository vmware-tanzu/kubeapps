// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import actions from "actions";
import * as ReactRedux from "react-redux";
import { NavLink } from "react-router-dom";
import { getStore, initialState, mountWrapper } from "shared/specs/mountWrapper";
import { IStoreState } from "shared/types";
import { app } from "shared/url";
import Header from "./Header";

let spyOnUseDispatch: jest.SpyInstance;
beforeEach(() => {
  actions.namespace = {
    ...actions.namespace,
    fetchNamespaces: jest.fn(),
    canCreate: jest.fn(),
  };
  const mockDispatch = jest.fn(res => res);
  spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
});

afterEach(() => {
  spyOnUseDispatch.mockRestore();
  jest.restoreAllMocks();
});

const defaultState = {
  ...initialState,
  clusters: {
    ...initialState.clusters,
    currentCluster: "default",
    clusters: {
      ...initialState.clusters.clusters,
      default: {
        ...initialState.clusters.clusters[initialState.clusters.currentCluster],
        currentNamespace: "default",
        namespaces: ["default", "other"],
      },
    },
  },
  auth: { ...initialState.auth, authenticated: true },
  config: { ...initialState.config, appVersion: "v2.0.0" },
} as IStoreState;

it("fetch namespaces and the ability to create them", () => {
  mountWrapper(getStore(defaultState), <Header />);
  expect(actions.namespace.fetchNamespaces).toHaveBeenCalled();
  expect(actions.namespace.canCreate).toHaveBeenCalled();
});

it("renders the header links and titles", () => {
  const wrapper = mountWrapper(getStore(defaultState), <Header />);
  const items = wrapper.find(".header-nav").find(NavLink);
  const expectedItems = [
    { children: "Applications", to: app.apps.list("default", "default") },
    { children: "Catalog", to: app.catalog("default", "default") },
  ];
  expect(items.length).toEqual(expectedItems.length);
  expectedItems.forEach((expectedItem, index) => {
    expect(expectedItem.children).toBe(items.at(index).text());
    expect(expectedItem.to).toBe(items.at(index).prop("to"));
  });
});

it("should skip the links if it's not authenticated", () => {
  const wrapper = mountWrapper(
    getStore({
      ...defaultState,
      auth: { authenticated: false },
    } as Partial<IStoreState>),
    <Header />,
  );
  const items = wrapper.find(".nav-link");
  expect(items).not.toExist();
});

it("should skip the links if the namespace info is not available", () => {
  const wrapper = mountWrapper(
    getStore({
      ...defaultState,
      clusters: {
        ...initialState.clusters,
        currentCluster: "default",
        clusters: {
          ...initialState.clusters.clusters,
          default: {
            ...initialState.clusters.clusters[initialState.clusters.currentCluster],
            currentNamespace: "",
            namespaces: [],
          },
        },
      },
    } as Partial<IStoreState>),
    <Header />,
  );
  const items = wrapper.find(".nav-link");
  expect(items).not.toExist();
});
