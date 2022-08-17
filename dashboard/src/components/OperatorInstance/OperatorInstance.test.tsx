// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import actions from "actions";
import AccessURLTable from "components/AppView/AccessURLTable/AccessURLTable";
import AppNotes from "components/AppView/AppNotes/AppNotes";
import AppSecrets from "components/AppView/AppSecrets";
import AppValues from "components/AppView/AppValues/AppValues";
import ResourceTabs from "components/AppView/ResourceTabs";
import ConfirmDialog from "components/ConfirmDialog/ConfirmDialog";
import Alert from "components/js/Alert";
import LoadingWrapper from "components/LoadingWrapper/LoadingWrapper";
import OperatorHeader from "components/OperatorView/OperatorHeader";
import ApplicationStatusContainer from "containers/ApplicationStatusContainer";
import { act } from "react-dom/test-utils";
import * as ReactRedux from "react-redux";
import { defaultStore, getStore, initialState, mountWrapper } from "shared/specs/mountWrapper";
import { FetchError, IStoreState } from "shared/types";
import OperatorInstance from "./OperatorInstance";

const defaultProps = {
  csvName: "foo",
  crdName: "foo.kubeapps.com",
  cluster: initialState.config.kubeappsCluster,
  namespace: "kubeapps",
  instanceName: "bar",
};

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
const kubeaActions = { ...actions.operators };
beforeEach(() => {
  actions.operators = {
    ...actions.operators,
    getCSV: jest.fn(),
    getResource: jest.fn(),
  };
  const mockDispatch = jest.fn(res => res);
  spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
});

afterEach(() => {
  actions.operators = { ...kubeaActions };
  spyOnUseDispatch.mockRestore();
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
    <OperatorInstance {...defaultProps} />,
  );
  expect(wrapper.find(Alert)).toIncludeText("Boom!");
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
    <OperatorInstance {...defaultProps} />,
  );
  expect(wrapper.find(Alert)).toIncludeText("Boom!");
});

it("renders an delete error", () => {
  const wrapper = mountWrapper(
    getStore({
      operators: {
        csv: defaultCSV,
        errors: { resource: { update: new Error("Boom!") } },
      },
    } as Partial<IStoreState>),
    <OperatorInstance {...defaultProps} />,
  );
  expect(wrapper.find(Alert)).toIncludeText("Boom!");
});

it("retrieves CSV and resource when mounted", () => {
  const getCSV = jest.fn();
  const getResource = jest.fn();
  actions.operators.getCSV = getCSV;
  actions.operators.getResource = getResource;
  mountWrapper(defaultStore, <OperatorInstance {...defaultProps} />);
  expect(getCSV).toHaveBeenCalledWith(
    defaultProps.cluster,
    defaultProps.namespace,
    defaultProps.csvName,
  );
  expect(getResource).toHaveBeenCalledWith(
    defaultProps.cluster,
    defaultProps.namespace,
    defaultProps.csvName,
    defaultProps.crdName,
    defaultProps.instanceName,
  );
});

it("renders a loading wrapper", () => {
  const wrapper = mountWrapper(
    getStore({ operators: { isFetching: true } } as Partial<IStoreState>),
    <OperatorInstance {...defaultProps} />,
  );
  expect(wrapper.find(LoadingWrapper)).toExist();
});

it("renders all the subcomponents", () => {
  const wrapper = mountWrapper(
    getStore({ operators: { csv: defaultCSV, resource } } as Partial<IStoreState>),
    <OperatorInstance {...defaultProps} />,
  );
  expect(wrapper.find(ApplicationStatusContainer)).toExist();
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
    <OperatorInstance {...defaultProps} />,
  );
  expect(wrapper.find(AppNotes)).not.toExist();
  expect(wrapper.find(AppValues)).not.toExist();
});

it("deletes the resource", async () => {
  const deleteResource = jest.fn().mockReturnValue(true);
  actions.operators.deleteResource = deleteResource;
  const wrapper = mountWrapper(
    getStore({ operators: { csv: defaultCSV, resource } } as Partial<IStoreState>),
    <OperatorInstance {...defaultProps} />,
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
  expect(deleteResource).toHaveBeenCalledWith(
    defaultProps.cluster,
    defaultProps.namespace,
    "foo",
    resource,
  );
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
    <OperatorInstance {...defaultProps} />,
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
    <OperatorInstance {...defaultProps} />,
  );
  const resources = wrapper.find(ResourceTabs).props();
  const resourcesKeys = Object.keys(resources).filter(k => k !== "otherResources");
  resourcesKeys.forEach(k => expect(resources[k].length).toBe(1));
});
