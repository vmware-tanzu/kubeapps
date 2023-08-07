// Copyright 2020-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import actions from "actions";
import AlertGroup from "components/AlertGroup";
import * as ReactRedux from "react-redux";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { IClusterState } from "reducers/cluster";
import { IOperatorsState } from "reducers/operators";
import { getStore, initialState, mountWrapper } from "shared/specs/mountWrapper";
import { IStoreState } from "shared/types";
import OperatorDescription from "./OperatorDescription";
import OperatorView from "./OperatorView";

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
    getCSV: jest.fn(),
  };
  const mockDispatch = jest.fn();
  spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
});

afterEach(() => {
  actions.operators = { ...kubeActions };
  spyOnUseDispatch.mockRestore();
});

it("calls getOperator when mounting the component", () => {
  const getOperator = jest.fn();
  actions.operators.getOperator = getOperator;
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

  mountWrapper(
    store,
    <MemoryRouter initialEntries={["/c/default/ns/default/operators/foo"]}>
      <Routes>
        <Route path={"/c/:cluster/ns/:namespace/operators/:operator"} element={<OperatorView />} />
      </Routes>
    </MemoryRouter>,
    false,
  );

  expect(getOperator).toHaveBeenCalledWith("default-cluster", "kubeapps", "foo");
});

it("tries to get the CSV for the current operator", () => {
  const getCSV = jest.fn();
  actions.operators.getCSV = getCSV;
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

  mountWrapper(store, <OperatorView />);

  expect(getCSV).toHaveBeenCalledWith("default-cluster", "kubeapps", "foo.1.0.0");
});

it("shows an error if it exists", () => {
  const wrapper = mountWrapper(
    getStore({
      operators: { errors: { operator: { fetch: new Error("boom") } } },
    } as Partial<IStoreState>),
    <OperatorView />,
  );
  expect(wrapper.find(AlertGroup)).toIncludeText("boom");
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
    <MemoryRouter initialEntries={["/c/default/ns/default/operators/foo"]}>
      <Routes>
        <Route path={"/c/:cluster/ns/:namespace/operators/:operator"} element={<OperatorView />} />
      </Routes>
    </MemoryRouter>,
    false,
  );
  expect(wrapper.find(AlertGroup)).toIncludeText(
    "Operator foo doesn't define a valid channel. This is needed to extract required info",
  );
});

it("selects the default channel", () => {
  const operator = {
    ...defaultOperator,
    status: {
      ...defaultOperator.status,
      channels: [{ name: "alpha" }, defaultOperator.status.channels[0]],
    },
  };
  const wrapper = mountWrapper(
    getStore({ operators: { operator } } as Partial<IStoreState>),
    <OperatorView />,
  );
  expect(wrapper.find(OperatorDescription).prop("description")).toEqual(
    "this is a testing operator",
  );
});

it("disables the Header deploy button if the subscription already exists", () => {
  const wrapper = mountWrapper(
    getStore({
      operators: {
        operator: defaultOperator,
        subscriptions: [{ spec: { name: defaultOperator.metadata.name } }],
      },
    } as Partial<IStoreState>),
    <OperatorView />,
  );
  wrapper.find(CdsButton).forEach(button => expect(button).toBeDisabled());
});
