// Copyright 2020-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import actions from "actions";
import Alert from "components/js/Alert";
import * as ReactRedux from "react-redux";
import { defaultStore, getStore, initialState, mountWrapper } from "shared/specs/mountWrapper";
import { IStoreState } from "shared/types";
import OperatorNew from "./OperatorNew";
import { IOperatorsState } from "reducers/operators";
import { IClusterState } from "reducers/cluster";
import { MemoryRouter, Route } from "react-router-dom";

const defaultOperator = {
  metadata: {
    name: "foo",
    namespace: "kubeapps",
  },
  status: {
    provider: {
      name: "Kubeapps",
    },
    defaultChannel: "beta",
    channels: [
      {
        name: "beta",
        currentCSV: "foo.1.0.0",
        currentCSVDesc: {
          displayName: "Foo",
          version: "1.0.0",
          description: "this is a testing operator",
          annotations: {
            capabilities: "Basic Install",
            repository: "github.com/vmware-tanzu/kubeapps",
            containerImage: "kubeapps/kubeapps",
            createdAt: "one day",
          },
          installModes: [],
        },
      },
    ],
  },
} as any;

let spyOnUseDispatch: jest.SpyInstance;
const kubeActions = { ...actions.operators };
beforeEach(() => {
  actions.operators = {
    ...actions.operators,
    getOperator: jest.fn(),
  };
  const mockDispatch = jest.fn(res => res);
  spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
});

afterEach(() => {
  actions.operators = { ...kubeActions };
  spyOnUseDispatch.mockRestore();
});

it("calls getOperator when mounting the component", () => {
  const getOperator = jest.fn();
  actions.operators.getOperator = getOperator;

  mountWrapper(
    defaultStore,
    <MemoryRouter initialEntries={["/c/default/ns/default/operators/new/foo"]}>
      <Route path={"/c/:cluster/ns/:namespace/operators/new/:operator"}>
        <OperatorNew />
      </Route>
    </MemoryRouter>,
  );
  expect(getOperator).toHaveBeenCalledWith(
    initialState.clusters.currentCluster,
    initialState.clusters.clusters[initialState.clusters.currentCluster].currentNamespace,
    "foo",
  );
});

it("parses the default channel when receiving the operator", () => {
  const wrapper = mountWrapper(
    getStore({ operators: { operator: defaultOperator } } as Partial<IStoreState>),
    <OperatorNew />,
  );
  const input = wrapper.find("#operator-channel-beta");
  expect(input).toExist();
  expect(input).toBeChecked();
});

it("renders a fetch error if present", () => {
  const wrapper = mountWrapper(
    getStore({
      operators: { errors: { operator: { fetch: new Error("Boom") } } },
    } as Partial<IStoreState>),
    <OperatorNew />,
  );
  expect(wrapper.find(Alert)).toIncludeText("Boom");
});

it("renders a create error if present", () => {
  const wrapper = mountWrapper(
    getStore({
      operators: { errors: { operator: { create: new Error("Boom") } } },
    } as Partial<IStoreState>),
    <OperatorNew />,
  );
  expect(wrapper.find(Alert)).toIncludeText("Boom");
});

it("shows an error if the operator doesn't have any channel defined", () => {
  const operator = {
    ...initialState.operators.operator,
    status: {
      ...initialState.operators.operator?.status,
      channels: [],
    },
  };
  const store = getStore({
    ...initialState,
    operators: { ...initialState.operators, operator },
  } as Partial<IStoreState>);

  const wrapper = mountWrapper(
    store,
    <MemoryRouter initialEntries={["/c/default/ns/default/operators/new/foo"]}>
      <Route path={"/c/:cluster/ns/:namespace/operators/new/:operator"}>
        <OperatorNew />
      </Route>
    </MemoryRouter>,
  );

  expect(wrapper.find(Alert)).toIncludeText(
    "Operator foo doesn't define a valid channel. This is needed to extract required info",
  );
});

it("disables the submit button if the operators ns is selected", () => {
  const store = getStore({
    operators: { operator: defaultOperator } as Partial<IOperatorsState>,
    clusters: {
      currentCluster: "default-cluster",
      clusters: {
        "default-cluster": {
          currentNamespace: "operators",
        } as Partial<IClusterState>,
      },
    },
  } as Partial<IStoreState>);
  const wrapper = mountWrapper(store, <OperatorNew />);
  expect(wrapper.find(CdsButton)).toBeDisabled();
  expect(wrapper.find(Alert)).toIncludeText(
    'It\'s not possible to install a namespaced operator in the "operators" namespace',
  );
});

it("deploys an operator", async () => {
  const createOperator = jest.fn().mockReturnValue(true);
  actions.operators.createOperator = createOperator;
  const store = getStore({
    operators: { operator: defaultOperator } as Partial<IOperatorsState>,
    clusters: {
      currentCluster: "default-cluster",
      clusters: {
        "default-cluster": {
          currentNamespace: "kubeapps",
        } as Partial<IClusterState>,
      },
    },
  } as Partial<IStoreState>);

  const wrapper = mountWrapper(store, <OperatorNew />);
  const onSubmit = wrapper.find("form").prop("onSubmit") as () => Promise<void>;
  await onSubmit();

  expect(createOperator).toHaveBeenCalledWith(
    initialState.clusters.currentCluster,
    "kubeapps",
    "foo",
    "beta",
    "Automatic",
    "foo.1.0.0",
  );
});
