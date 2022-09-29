// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import actions from "actions";
import Alert from "components/js/Alert";
import OperatorInstanceFormBody from "components/OperatorInstanceFormBody/OperatorInstanceFormBody";
import OperatorHeader from "components/OperatorView/OperatorHeader";
import * as ReactRedux from "react-redux";
import { defaultStore, getStore, initialState, mountWrapper } from "shared/specs/mountWrapper";
import { FetchError, IStoreState } from "shared/types";
import OperatorInstanceUpdateForm, {
  IOperatorInstanceUpgradeFormProps,
} from "./OperatorInstanceUpdateForm";

const defaultProps: IOperatorInstanceUpgradeFormProps = {
  csvName: "foo",
  crdName: "foo-cluster",
  cluster: initialState.config.kubeappsCluster,
  namespace: "kubeapps",
  resourceName: "my-foo",
};

const defaultResource = {
  kind: "Foo",
  apiVersion: "v1",
  metadata: {
    name: "my-foo",
  },
} as any;

const defaultCRD = {
  name: defaultProps.crdName,
  kind: "Foo",
  description: "useful description",
} as any;

const defaultCSV = {
  metadata: {
    annotations: {
      "alm-examples": '[{"kind": "Foo", "apiVersion": "v1"}]',
    },
  },
  spec: {
    customresourcedefinitions: {
      owned: [defaultCRD],
    },
  },
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

  // mock the window.ResizeObserver, required by the MonacoEditor for the layout
  Object.defineProperty(window, "ResizeObserver", {
    writable: true,
    configurable: true,
    value: jest.fn().mockImplementation(() => ({
      observe: jest.fn(),
      unobserve: jest.fn(),
      disconnect: jest.fn(),
    })),
  });

  // mock the window.HTMLCanvasElement.getContext(), required by the MonacoEditor for the layout
  Object.defineProperty(HTMLCanvasElement.prototype, "getContext", {
    writable: true,
    configurable: true,
    value: jest.fn().mockImplementation(() => ({
      clearRect: jest.fn(),
    })),
  });
});

afterEach(() => {
  actions.operators = { ...kubeaActions };
  spyOnUseDispatch.mockRestore();
  jest.restoreAllMocks();
});

it("gets resource and CSV", () => {
  const getResource = jest.fn();
  const getCSV = jest.fn();
  actions.operators.getResource = getResource;
  actions.operators.getCSV = getCSV;
  mountWrapper(defaultStore, <OperatorInstanceUpdateForm {...defaultProps} />);
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
    defaultProps.resourceName,
  );
});

it("set default and deployed values", () => {
  const wrapper = mountWrapper(
    getStore({
      operators: {
        resource: defaultResource,
        csv: defaultCSV,
      },
    } as Partial<IStoreState>),
    <OperatorInstanceUpdateForm {...defaultProps} />,
  );
  expect(wrapper.find(OperatorInstanceFormBody).props()).toMatchObject({
    defaultValues: "kind: Foo\napiVersion: v1\n",
    deployedValues: "kind: Foo\napiVersion: v1\nmetadata:\n  name: my-foo\n",
  });
});

it("renders an error if the resource is not populated", () => {
  const wrapper = mountWrapper(defaultStore, <OperatorInstanceUpdateForm {...defaultProps} />);
  expect(wrapper.find(Alert)).toIncludeText("Resource my-foo not found");
});

it("renders only an error if the resource is not found", () => {
  const wrapper = mountWrapper(
    getStore({
      ...initialState,
      operators: {
        ...initialState.operators,
        errors: {
          ...initialState.operators.errors,
          fetch: new FetchError("not found"),
        },
      },
    } as Partial<IStoreState>),
    <OperatorInstanceUpdateForm {...defaultProps} />,
  );
  expect(wrapper.find(Alert)).toIncludeText("not found");
  expect(wrapper.find(OperatorHeader)).not.toExist();
});

it("should submit the form", () => {
  const updateResource = jest.fn();
  actions.operators.updateResource = updateResource;
  const wrapper = mountWrapper(
    getStore({
      operators: {
        resource: defaultResource,
        csv: defaultCSV,
      },
    } as Partial<IStoreState>),
    <OperatorInstanceUpdateForm {...defaultProps} />,
  );

  const form = wrapper.find("form");
  form.simulate("submit", { preventDefault: jest.fn() });

  expect(updateResource).toHaveBeenCalledWith(
    defaultProps.cluster,
    defaultProps.namespace,
    defaultResource.apiVersion,
    defaultProps.crdName,
    defaultProps.resourceName,
    defaultResource,
  );
});
