import { CdsButton } from "@clr/react/button";
import { CdsModal } from "@clr/react/modal";
import actions from "actions";
import Alert from "components/js/Alert";
import { cloneDeep } from "lodash";
import * as React from "react";
import { act } from "react-dom/test-utils";
import * as ReactRedux from "react-redux";
import { IClustersState } from "reducers/cluster";
import { definedNamespaces } from "shared/Namespace";
import { defaultStore, getStore, initialState, mountWrapper } from "shared/specs/mountWrapper";
import ContextSelector from "./ContextSelector";

let spyOnUseDispatch: jest.SpyInstance;
const kubeaActions = { ...actions.operators };
beforeEach(() => {
  actions.namespace = {
    ...actions.namespace,
    fetchNamespaces: jest.fn(),
    getNamespace: jest.fn(),
    setNamespace: jest.fn(),
    createNamespace: jest.fn(),
  };
  const mockDispatch = jest.fn(res => res);
  spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
});

afterEach(() => {
  actions.operators = { ...kubeaActions };
  spyOnUseDispatch.mockRestore();
});

it("gets a namespace", () => {
  const getNamespace = jest.fn();
  actions.namespace.getNamespace = getNamespace;
  mountWrapper(defaultStore, <ContextSelector />);

  expect(getNamespace).toHaveBeenCalledWith(
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
    (wrapper
      .find(CdsButton)
      .filterWhere(b => b.text() === "Change Context")
      .prop("onClick") as any)();
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
  const wrapper = mountWrapper(getStore({ clusters }), <ContextSelector />);
  expect(
    wrapper
      .find("select")
      .at(0)
      .prop("value"),
  ).toBe("bar");
});

it("shows the current namespace", () => {
  const clusters = cloneDeep(initialState.clusters);
  clusters.clusters[clusters.currentCluster].currentNamespace = "other";
  const wrapper = mountWrapper(getStore({ clusters }), <ContextSelector />);
  expect(
    wrapper
      .find("select")
      .at(1)
      .prop("value"),
  ).toBe("other");
});

it("includes all namespaces", () => {
  const clusters = cloneDeep(initialState.clusters);
  clusters.clusters[clusters.currentCluster].currentNamespace = definedNamespaces.all;
  const wrapper = mountWrapper(getStore({ clusters }), <ContextSelector />);
  expect(wrapper.find("label").findWhere(l => l.prop("htmlFor") === "namespaces")).toIncludeText(
    "All Namespaces",
  );
  expect(
    wrapper
      .find("select")
      .find("option")
      .filterWhere(op => op.prop("value") === definedNamespaces.all),
  ).toExist();
});

it("submits the form to create a new namespace", () => {
  const createNamespace = jest.fn();
  actions.namespace.createNamespace = createNamespace;
  const wrapper = mountWrapper(defaultStore, <ContextSelector />);

  const modalButton = wrapper.find(".flat-btn");
  act(() => {
    (modalButton.prop("onClick") as any)();
  });
  wrapper.update();
  expect(wrapper.find(CdsModal)).toHaveProp("hidden", false);

  act(() => {
    wrapper.find("input").simulate("change", { target: { value: "new-ns" } });
  });
  wrapper.update();

  act(() => {
    wrapper.find("form").simulate("submit", { preventDefault: jest.fn() });
  });
  wrapper.update();

  expect(createNamespace).toHaveBeenCalledWith(initialState.clusters.currentCluster, "new-ns");
});

it("shows an error creating a namespace", () => {
  const clusters = cloneDeep(initialState.clusters);
  clusters.clusters[clusters.currentCluster].error = { error: new Error("Boom"), action: "create" };

  const wrapper = mountWrapper(getStore({ clusters }), <ContextSelector />);
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
  const wrapper = mountWrapper(getStore({ clusters }), <ContextSelector />);
  expect(wrapper.find(".flat-btn")).toBeDisabled();
});
