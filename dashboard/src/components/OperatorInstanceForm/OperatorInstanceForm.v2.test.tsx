import actions from "actions";
import AdvancedDeploymentForm from "components/DeploymentFormBody/AdvancedDeploymentForm.v2";
import Alert from "components/js/Alert";
import OperatorInstanceFormBody from "components/OperatorInstanceFormBody/OperatorInstanceFormBody.v2";
import OperatorNotSupported from "components/OperatorList/OperatorsNotSupported.v2";
import * as React from "react";
import { act } from "react-dom/test-utils";
import * as ReactRedux from "react-redux";
import { defaultStore, getStore, initialState, mountWrapper } from "shared/specs/mountWrapper";
import { IClusterServiceVersion } from "../../shared/types";
import OperatorInstanceForm, { IOperatorInstanceFormProps } from "./OperatorInstanceForm.v2";

const defaultProps: IOperatorInstanceFormProps = {
  csvName: "foo",
  crdName: "foo-cluster",
  cluster: initialState.config.kubeappsCluster,
  namespace: "kubeapps",
};

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
  };
  const mockDispatch = jest.fn(res => res);
  spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
});

afterEach(() => {
  actions.operators = { ...kubeaActions };
  spyOnUseDispatch.mockRestore();
});

it("displays an alert if rendered for an additional cluster", () => {
  const props = { ...defaultProps, cluster: "other-cluster" };
  const wrapper = mountWrapper(defaultStore, <OperatorInstanceForm {...props} />);
  expect(wrapper.find(OperatorNotSupported)).toExist();
});

it("renders a fetch error", () => {
  const wrapper = mountWrapper(
    getStore({
      operators: {
        errors: { csv: { fetch: new Error("Boom!") } },
      },
    }),
    <OperatorInstanceForm {...defaultProps} />,
  );
  expect(wrapper.find(Alert)).toIncludeText("Boom!");
});

it("renders a create error", () => {
  const wrapper = mountWrapper(
    getStore({
      operators: {
        csv: defaultCSV,
        errors: { resource: { create: new Error("Boom!") } },
      },
    }),
    <OperatorInstanceForm {...defaultProps} />,
  );
  expect(wrapper.find(Alert)).toIncludeText("Boom!");
});

it("retrieves CSV when mounted", () => {
  const getCSV = jest.fn();
  actions.operators.getCSV = getCSV;
  mountWrapper(defaultStore, <OperatorInstanceForm {...defaultProps} />);
  expect(getCSV).toHaveBeenCalledWith(
    defaultProps.cluster,
    defaultProps.namespace,
    defaultProps.csvName,
  );
});

it("retrieves the example values and the target CRD from the given CSV", () => {
  const wrapper = mountWrapper(
    getStore({ operators: { csv: defaultCSV } }),
    <OperatorInstanceForm {...defaultProps} />,
  );
  expect(wrapper.find(OperatorInstanceFormBody).props()).toMatchObject({
    defaultValues: "kind: Foo\napiVersion: v1\n",
  });
});

it("defaults to empty defaultValues if the examples annotation is not found", () => {
  const csv = {
    ...defaultCSV,
    metadata: {},
  } as IClusterServiceVersion;
  const wrapper = mountWrapper(
    getStore({ operators: { csv } }),
    <OperatorInstanceForm {...defaultProps} />,
  );
  expect(wrapper.find(OperatorInstanceFormBody).props()).toMatchObject({
    defaultValues: "",
  });
});

it("renders an error if the CRD is not populated", () => {
  const wrapper = mountWrapper(defaultStore, <OperatorInstanceForm {...defaultProps} />);
  expect(wrapper.find(Alert)).toIncludeText("not found in the definition");
});

it("should submit the form", () => {
  const createResource = jest.fn();
  actions.operators.createResource = createResource;
  const wrapper = mountWrapper(
    getStore({ operators: { csv: defaultCSV } }),
    <OperatorInstanceForm {...defaultProps} />,
  );

  act(() => {
    (wrapper.find(AdvancedDeploymentForm).prop("handleValuesChange") as any)(
      "apiVersion: v1\nmetadata:\n  name: foo",
    );
  });
  wrapper.update();

  const form = wrapper.find("form");
  form.simulate("submit", { preventDefault: jest.fn() });

  const resource = {
    apiVersion: "v1",
    metadata: {
      name: "foo",
    },
  };
  expect(createResource).toHaveBeenCalledWith(
    defaultProps.cluster,
    defaultProps.namespace,
    resource.apiVersion,
    defaultCRD.name,
    resource,
  );
});
