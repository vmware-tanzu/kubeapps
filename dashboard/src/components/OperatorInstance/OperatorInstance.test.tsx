// Copyright 2020-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import { act } from "@testing-library/react";
import actions from "actions";
import AlertGroup from "components/AlertGroup";
import AccessURLTable from "components/AppView/AccessURLTable/AccessURLTable";
import AppNotes from "components/AppView/AppNotes/AppNotes";
import AppSecrets from "components/AppView/AppSecrets";
import AppValues from "components/AppView/AppValues/AppValues";
import ResourceTabs from "components/AppView/ResourceTabs";
import ApplicationStatus from "components/ApplicationStatus/ApplicationStatus";
import ConfirmDialog from "components/ConfirmDialog/ConfirmDialog";
import LoadingWrapper from "components/LoadingWrapper/LoadingWrapper";
import OperatorHeader from "components/OperatorView/OperatorHeader";
import * as ReactRedux from "react-redux";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { IClusterState } from "reducers/cluster";
import { IOperatorsState } from "reducers/operators";
import { getStore, initialState, mountWrapper } from "shared/specs/mountWrapper";
import { FetchError, IStoreState } from "shared/types";
import OperatorInstance from "./OperatorInstance";

const defaultCSV = {
  metadata: { name: "foo" },
  spec: {
    icon: [{}],
    customresourcedefinitions: {
      owned: [
        {
          name: "foo.kubeapps.com",
          version: "v1alpha1",
          kind: "Foo",
          resources: [{ kind: "Deployment" }],
        },
      ],
    },
  },
} as any;

const resource = {
  kind: "Foo",
  metadata: { name: "foo-instance" },
  spec: { test: true },
  status: { alive: true },
} as any;

let spyOnUseDispatch: jest.SpyInstance;
const kubeActions = { ...actions.operators };
beforeEach(() => {
  actions.operators = {
    ...actions.operators,
    getCSV: jest.fn(),
    getResource: jest.fn(),
  };
  const mockDispatch = jest.fn(res => res);
  spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);

  // mock the window.matchMedia for selecting the theme
  Object.defineProperty(window, "matchMedia", {
    writable: true,
    configurable: true,
    value: jest.fn().mockImplementation(query => ({
      matches: false,
      media: query,
      onchange: null,
      addListener: jest.fn(),
      removeListener: jest.fn(),
      addEventListener: jest.fn(),
      removeEventListener: jest.fn(),
      dispatchEvent: jest.fn(),
    })),
  });

  // mock the window.ResizeObserver, required by the MonacoDiffEditor for the layout
  Object.defineProperty(window, "ResizeObserver", {
    writable: true,
    configurable: true,
    value: jest.fn().mockImplementation(() => ({
      observe: jest.fn(),
      unobserve: jest.fn(),
      disconnect: jest.fn(),
    })),
  });

  // mock the window.HTMLCanvasElement.getContext(), required by the MonacoDiffEditor for the layout
  Object.defineProperty(HTMLCanvasElement.prototype, "getContext", {
    writable: true,
    configurable: true,
    value: jest.fn().mockImplementation(() => ({
      clearRect: jest.fn(),
      fillRect: jest.fn(),
    })),
  });
});

afterEach(() => {
  actions.operators = { ...kubeActions };
  spyOnUseDispatch.mockRestore();
  jest.restoreAllMocks();
});

it("renders a fetch error", () => {
  const wrapper = mountWrapper(
    getStore({
      ...initialState,
      operators: {
        ...initialState.operators,
        errors: { ...initialState.operators.errors, resource: { fetch: new FetchError("Boom!") } },
      },
    } as Partial<IStoreState>),
    <OperatorInstance />,
  );
  expect(wrapper.find(AlertGroup)).toIncludeText("Boom!");
  expect(wrapper.find(OperatorHeader)).not.toExist();
});

it("renders an update error", () => {
  const wrapper = mountWrapper(
    getStore({
      operators: {
        csv: defaultCSV,
        errors: { resource: { update: new Error("Boom!") } },
      },
    } as Partial<IStoreState>),
    <OperatorInstance />,
  );
  expect(wrapper.find(AlertGroup)).toIncludeText("Boom!");
});

it("renders an delete error", () => {
  const wrapper = mountWrapper(
    getStore({
      operators: {
        csv: defaultCSV,
        errors: { resource: { update: new Error("Boom!") } },
      },
    } as Partial<IStoreState>),
    <OperatorInstance />,
  );
  expect(wrapper.find(AlertGroup)).toIncludeText("Boom!");
});

it("retrieves CSV and resource when mounted", () => {
  const getCSV = jest.fn();
  const getResource = jest.fn();
  actions.operators.getCSV = getCSV;
  actions.operators.getResource = getResource;
  const store = getStore({
    operators: { csv: defaultCSV, resource } as Partial<IOperatorsState>,
    clusters: {
      currentCluster: "default-cluster",
      clusters: {
        "default-cluster": {
          currentNamespace: "kubeapps",
        } as Partial<IClusterState>,
      },
    },
  } as Partial<IStoreState>);
  mountWrapper(
    store,
    <MemoryRouter
      initialEntries={["/c/default/ns/default/operators-instances/foo/foo.kubeapps.com/bar"]}
    >
      <Routes>
        <Route
          path={"/c/:cluster/ns/:namespace/operators-instances/:csv/:crd/:instanceName"}
          element={<OperatorInstance />}
        />
      </Routes>
    </MemoryRouter>,
    false,
  );
  expect(getCSV).toHaveBeenCalledWith("default-cluster", "kubeapps", "foo");
  expect(getResource).toHaveBeenCalledWith(
    "default-cluster",
    "kubeapps",
    "foo",
    "foo.kubeapps.com",
    "bar",
  );
});

it("renders a loading wrapper", () => {
  const wrapper = mountWrapper(
    getStore({ operators: { isFetching: true } } as Partial<IStoreState>),
    <OperatorInstance />,
  );
  expect(wrapper.find(LoadingWrapper)).toExist();
});

it("renders all the subcomponents", () => {
  const wrapper = mountWrapper(
    getStore({ operators: { csv: defaultCSV, resource } } as Partial<IStoreState>),
    <OperatorInstance />,
  );
  expect(wrapper.find(ApplicationStatus)).toExist();
  expect(wrapper.find(AccessURLTable)).toExist();
  expect(wrapper.find(AppSecrets)).toExist();
  expect(wrapper.find(AppNotes)).toExist();
  expect(wrapper.find(ResourceTabs)).toExist();
  expect(wrapper.find(AppValues)).toExist();
});

it("skips AppNotes and AppValues if the resource doesn't have spec or status", () => {
  const wrapper = mountWrapper(
    getStore({
      operators: { csv: defaultCSV, resource: { ...resource, spec: undefined, status: undefined } },
    } as Partial<IStoreState>),
    <OperatorInstance />,
  );
  expect(wrapper.find(AppNotes)).not.toExist();
  expect(wrapper.find(AppValues)).not.toExist();
});

it("deletes the resource", async () => {
  const deleteResource = jest.fn().mockReturnValue(true);
  actions.operators.deleteResource = deleteResource;
  const store = getStore({
    operators: { csv: defaultCSV, resource } as Partial<IOperatorsState>,
    clusters: {
      currentCluster: "default-cluster",
      clusters: {
        "default-cluster": {
          currentNamespace: "kubeapps",
        } as Partial<IClusterState>,
      },
    },
  } as Partial<IStoreState>);
  const wrapper = mountWrapper(
    store,
    <MemoryRouter
      initialEntries={["/c/default/ns/default/operators-instances/foo/foo.kubeapps.com/bar"]}
    >
      <Routes>
        <Route
          path={"/c/:cluster/ns/:namespace/operators-instances/:csv/:crd/:instanceName"}
          element={<OperatorInstance />}
        />
      </Routes>
    </MemoryRouter>,
    false,
  );

  act(() => {
    (
      wrapper
        .find(CdsButton)
        .filterWhere(b => b.text().includes("Delete"))
        .prop("onClick") as any
    )();
  });
  wrapper.update();
  const dialog = wrapper.find(ConfirmDialog);
  expect(dialog.prop("modalIsOpen")).toEqual(true);
  await act(async () => {
    await (dialog.prop("onConfirm") as any)();
  });
  expect(deleteResource).toHaveBeenCalledWith("default-cluster", "kubeapps", "foo", resource);
});

it("updates the state with the CRD resources", () => {
  const wrapper = mountWrapper(
    getStore({
      ...initialState,
      operators: { ...initialState.operators, csv: defaultCSV, resource },
      kube: {
        ...initialState.kube,
        kinds: { Foo: { apiVersion: "apps/v1", plural: "foos", namespaced: true } },
      },
    } as Partial<IStoreState>),
    <MemoryRouter
      initialEntries={["/c/default/ns/default/operators-instances/foo/foo.kubeapps.com/bar"]}
    >
      <Routes>
        <Route
          path={"/c/:cluster/ns/:namespace/operators-instances/:csv/:crd/:instanceName"}
          element={<OperatorInstance />}
        />
      </Routes>
    </MemoryRouter>,
    false,
  );
  expect(wrapper.find(ResourceTabs).prop("deployments")).toMatchObject([
    {
      apiVersion: "apps/v1",
      filter: {
        metadata: {
          ownerReferences: [
            {
              kind: "Foo",
              name: "foo-instance",
            },
          ],
        },
      },
    },
  ]);
});

it("updates the state with all the resources if the CRD doesn't define any", () => {
  const csvWithoutResource = {
    ...defaultCSV,
    spec: {
      ...defaultCSV.spec,
      customresourcedefinitions: {
        owned: [
          {
            name: "foo.kubeapps.com",
            version: "v1alpha1",
            kind: "Foo",
          },
        ],
      },
    },
  } as any;
  const wrapper = mountWrapper(
    getStore({
      ...initialState,
      operators: { ...initialState.operators, csv: csvWithoutResource, resource },
      kube: {
        ...initialState.kube,
        kinds: { Foo: { apiVersion: "apps/v1", plural: "foos", namespaced: true } },
      },
    } as Partial<IStoreState>),
    <MemoryRouter
      initialEntries={["/c/default/ns/default/operators-instances/foo/foo.kubeapps.com/bar"]}
    >
      <Routes>
        <Route
          path={"/c/:cluster/ns/:namespace/operators-instances/:csv/:crd/:instanceName"}
          element={<OperatorInstance />}
        />
      </Routes>
    </MemoryRouter>,
    false,
  );
  const resources: { [index: string]: any } = wrapper.find(ResourceTabs).props();
  const resourcesKeys = Object.keys(resources).filter(k => k !== "otherResources");
  resourcesKeys.forEach(k => expect(resources[k].length).toBe(1));
});
