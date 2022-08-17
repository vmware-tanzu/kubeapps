// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import actions from "actions";
import Alert from "components/js/Alert";
import * as ReactRedux from "react-redux";
import { defaultStore, getStore, initialState, mountWrapper } from "shared/specs/mountWrapper";
import { IStoreState } from "shared/types";
import OperatorDescription from "./OperatorDescription";
import OperatorView from "./OperatorView";

const defaultProps = {
  operatorName: "foo",
  cluster: initialState.config.kubeappsCluster,
  namespace: "kubeapps",
};

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
const kubeaActions = { ...actions.operators };
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
  actions.operators = { ...kubeaActions };
  spyOnUseDispatch.mockRestore();
});

it("calls getOperator when mounting the component", () => {
  const getOperator = jest.fn();
  actions.operators.getOperator = getOperator;
  mountWrapper(defaultStore, <OperatorView {...defaultProps} />);
  expect(getOperator).toHaveBeenCalledWith(
    defaultProps.cluster,
    defaultProps.namespace,
    defaultProps.operatorName,
  );
});

it("tries to get the CSV for the current operator", () => {
  const getCSV = jest.fn();
  actions.operators.getCSV = getCSV;
  mountWrapper(
    getStore({ operators: { operator: defaultOperator } } as Partial<IStoreState>),
    <OperatorView {...defaultProps} />,
  );

  expect(getCSV).toHaveBeenCalledWith(
    defaultProps.cluster,
    defaultOperator.metadata.namespace,
    defaultOperator.status.channels[0].currentCSV,
  );
});

it("shows an error if it exists", () => {
  const wrapper = mountWrapper(
    getStore({
      operators: { errors: { operator: { fetch: new Error("boom") } } },
    } as Partial<IStoreState>),
    <OperatorView {...defaultProps} />,
  );
  expect(wrapper.find(Alert)).toIncludeText("boom");
});

it("shows an error if the operator doesn't have any channel defined", () => {
  const operator = {
    ...initialState.operators.operator,
    status: {
      ...initialState.operators.operator?.status,
      channels: [],
    },
  };
  const wrapper = mountWrapper(
    getStore({
      ...initialState,
      operators: { ...initialState.operators, operator },
    } as Partial<IStoreState>),
    <OperatorView {...defaultProps} />,
  );
  expect(wrapper.find(Alert)).toIncludeText(
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
    <OperatorView {...defaultProps} />,
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
    <OperatorView {...defaultProps} />,
  );
  wrapper.find(CdsButton).forEach(button => expect(button).toBeDisabled());
});
