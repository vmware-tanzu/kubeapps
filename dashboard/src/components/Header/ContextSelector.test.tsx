// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import { CdsModal } from "@cds/react/modal";
import actions from "actions";
import Alert from "components/js/Alert";
import { createMemoryHistory } from "history";
import { cloneDeep } from "lodash";
import { act } from "react-dom/test-utils";
import * as ReactRedux from "react-redux";
import * as ReactRouter from "react-router";
import { Router } from "react-router-dom";
import { IClustersState } from "reducers/cluster";
import { defaultStore, getStore, initialState, mountWrapper } from "shared/specs/mountWrapper";
import { IStoreState } from "shared/types";
import ContextSelector from "./ContextSelector";

let spyOnUseDispatch: jest.SpyInstance;
let spyOnUseHistory: jest.SpyInstance;
const kubeaActions = { ...actions.operators };
beforeEach(() => {
  actions.namespace = {
    ...actions.namespace,
    fetchNamespaces: jest.fn(),
    checkNamespaceExists: jest.fn(),
    setNamespace: jest.fn(),
    createNamespace: jest.fn(),
  };
  const mockDispatch = jest.fn(res => res);
  spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
  spyOnUseHistory = jest
    .spyOn(ReactRouter, "useHistory")
    .mockReturnValue({ push: jest.fn() } as any);
});

afterEach(() => {
  actions.operators = { ...kubeaActions };
  spyOnUseDispatch.mockRestore();
  spyOnUseHistory.mockRestore();
});

it("gets a namespace", () => {
  const checkNamespaceExists = jest.fn();
  actions.namespace.checkNamespaceExists = checkNamespaceExists;
  mountWrapper(defaultStore, <ContextSelector />);

  expect(checkNamespaceExists).toHaveBeenCalledWith(
    initialState.clusters.currentCluster,
    initialState.clusters.clusters[initialState.clusters.currentCluster].currentNamespace,
  );
});

it("opens the dropdown menu", () => {
  const wrapper = mountWrapper(defaultStore, <ContextSelector />);
  expect(wrapper.find(".dropdown")).not.toHaveClassName("open");
  const menu = wrapper.find("button");
  menu.simulate("click");
  wrapper.update();
  expect(wrapper.find(".dropdown")).toHaveClassName("open");
});

it("selects a different namespace", () => {
  const setNamespace = jest.fn();
  actions.namespace = {
    ...actions.namespace,
    setNamespace,
  };
  const wrapper = mountWrapper(defaultStore, <ContextSelector />);
  wrapper
    .find("select")
    .findWhere(s => s.prop("name") === "namespaces")
    .simulate("change", { target: { value: "other" } });
  act(() => {
    (
      wrapper
        .find(CdsButton)
        .filterWhere(b => b.text() === "Change Context")
        .prop("onClick") as any
    )();
  });
  expect(setNamespace).toHaveBeenCalledWith(initialState.clusters.currentCluster, "other");
});

it("shows the current cluster", () => {
  const clusters = {
    currentCluster: "bar",
    clusters: {
      foo: {
        currentNamespace: "default",
        namespaces: ["default"],
        canCreateNS: true,
      },
      bar: {
        currentNamespace: "default",
        namespaces: ["default"],
        canCreateNS: true,
      },
    },
  } as IClustersState;
  const wrapper = mountWrapper(getStore({ clusters } as Partial<IStoreState>), <ContextSelector />);
  expect(wrapper.find("select").at(0).prop("value")).toBe("bar");
});

it("shows the current namespace", () => {
  const clusters = cloneDeep(initialState.clusters);
  clusters.clusters[clusters.currentCluster].currentNamespace = "other";
  const wrapper = mountWrapper(getStore({ clusters } as Partial<IStoreState>), <ContextSelector />);
  expect(wrapper.find("select").at(1).prop("value")).toBe("other");
});

it("submits the form to create a new namespace", () => {
  const createNamespace = jest.fn();
  actions.namespace.createNamespace = createNamespace;
  const wrapper = mountWrapper(defaultStore, <ContextSelector />);

  const modalButton = wrapper.find(".flat-btn").first();
  act(() => {
    (modalButton.prop("onClick") as any)();
  });
  wrapper.update();
  expect(wrapper.find(CdsModal)).toExist();

  act(() => {
    wrapper.find("input").simulate("change", { target: { value: "new-ns" } });
  });
  wrapper.update();

  act(() => {
    wrapper.find("form").simulate("submit", { preventDefault: jest.fn() });
  });
  wrapper.update();

  expect(createNamespace).toHaveBeenCalledWith(initialState.clusters.currentCluster, "new-ns", {});
});

it("submits the form to create a new namespace with custom labels", () => {
  const createNamespace = jest.fn();
  actions.namespace.createNamespace = createNamespace;

  const config = cloneDeep(initialState.config);
  config.createNamespaceLabels = {
    "managed-by": "kubeapps",
  };
  const wrapper = mountWrapper(getStore({ config } as Partial<IStoreState>), <ContextSelector />);

  const modalButton = wrapper.find(".flat-btn").first();
  act(() => {
    (modalButton.prop("onClick") as any)();
  });
  wrapper.update();
  expect(wrapper.find(CdsModal)).toExist();

  act(() => {
    wrapper.find("input").simulate("change", { target: { value: "new-ns" } });
  });
  wrapper.update();

  act(() => {
    wrapper.find("form").simulate("submit", { preventDefault: jest.fn() });
  });
  wrapper.update();

  expect(createNamespace).toHaveBeenCalledWith(initialState.clusters.currentCluster, "new-ns", {
    "managed-by": "kubeapps",
  });
});

it("shows an error creating a namespace", () => {
  const clusters = cloneDeep(initialState.clusters);
  clusters.clusters[clusters.currentCluster].error = { error: new Error("Boom"), action: "create" };

  const wrapper = mountWrapper(getStore({ clusters } as Partial<IStoreState>), <ContextSelector />);

  const modalButton = wrapper.find(".flat-btn").first();
  act(() => {
    (modalButton.prop("onClick") as any)();
  });
  wrapper.update();

  // The error will be within the modal
  expect(wrapper.find(CdsModal).find(Alert)).toExist();
});

it("disables the create button if not allowed", () => {
  const clusters = {
    currentCluster: "foo",
    clusters: {
      foo: {
        currentNamespace: "default",
        namespaces: ["default"],
        canCreateNS: false,
      },
    },
  } as IClustersState;
  const wrapper = mountWrapper(getStore({ clusters } as Partial<IStoreState>), <ContextSelector />);
  expect(wrapper.find(".flat-btn").first()).toBeDisabled();
});

it("disables the change context button if namespace is not loaded yet", () => {
  const clusters = {
    currentCluster: "foo",
    clusters: {
      foo: {
        currentNamespace: "default",
        namespaces: [],
        canCreateNS: false,
      },
    },
  } as IClustersState;
  const wrapper = mountWrapper(getStore({ clusters } as Partial<IStoreState>), <ContextSelector />);
  expect(wrapper.find(CdsButton).filterWhere(b => b.text() === "Change Context")).toBeDisabled();
});

it("changes the location with the new namespace", () => {
  const push = jest.fn();
  spyOnUseHistory = jest.spyOn(ReactRouter, "useHistory").mockReturnValue({ push } as any);
  const history = createMemoryHistory({ initialEntries: ["/c/default-cluster/ns/ns-bar/catalog"] });
  const wrapper = mountWrapper(
    defaultStore,
    <Router history={history}>
      <ContextSelector />
    </Router>,
  );
  wrapper
    .find("select")
    .findWhere(s => s.prop("name") === "namespaces")
    .simulate("change", { target: { value: "other" } });
  act(() => {
    (
      wrapper
        .find(CdsButton)
        .filterWhere(b => b.text() === "Change Context")
        .prop("onClick") as any
    )();
  });
  expect(history.location.pathname).toBe("/c/default-cluster/ns/other/catalog");
});

it("changes the location with the new cluster and namespace", () => {
  const history = createMemoryHistory({ initialEntries: ["/c/default-cluster/ns/ns-bar/catalog"] });
  const wrapper = mountWrapper(
    defaultStore,
    <Router history={history}>
      <ContextSelector />
    </Router>,
  );
  wrapper
    .find("select")
    .findWhere(s => s.prop("name") === "clusters")
    .simulate("change", { target: { value: "second-cluster" } });
  wrapper
    .find("select")
    .findWhere(s => s.prop("name") === "namespaces")
    .simulate("change", { target: { value: "other" } });
  act(() => {
    (
      wrapper
        .find(CdsButton)
        .filterWhere(b => b.text() === "Change Context")
        .prop("onClick") as any
    )();
  });
  expect(history.location.pathname).toBe("/c/second-cluster/ns/other/catalog");
});

it("don't call push if the pathname is not recognized", () => {
  const history = createMemoryHistory({ initialEntries: ["/foo"] });
  const wrapper = mountWrapper(
    defaultStore,
    <Router history={history}>
      <ContextSelector />
    </Router>,
  );
  wrapper
    .find("select")
    .findWhere(s => s.prop("name") === "namespaces")
    .simulate("change", { target: { value: "other" } });
  act(() => {
    (
      wrapper
        .find(CdsButton)
        .filterWhere(b => b.text() === "Change Context")
        .prop("onClick") as any
    )();
  });
  expect(history.location.pathname).toBe("/foo");
});
