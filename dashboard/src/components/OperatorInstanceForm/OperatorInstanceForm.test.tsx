import OperatorNotSupported from "components/OperatorList/OperatorsNotSupported";
import { mount, shallow } from "enzyme";
import * as React from "react";
import { IClusterServiceVersion } from "../../shared/types";
import NotFoundErrorPage from "../ErrorAlert/NotFoundErrorAlert";
import OperatorInstanceForm, { IOperatorInstanceFormProps } from "./OperatorInstanceForm";

const defaultProps: IOperatorInstanceFormProps = {
  csvName: "foo",
  crdName: "foo-cluster",
  isFetching: false,
  cluster: "default",
  namespace: "kubeapps",
  kubeappsCluster: "default",
  getCSV: jest.fn(),
  createResource: jest.fn(),
  push: jest.fn(),
  errors: {},
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

it("displays an alert if rendered for an additional cluster", () => {
  const props = { ...defaultProps, cluster: "other-cluster" };
  const wrapper = shallow(<OperatorInstanceForm {...props} />);
  expect(wrapper.find(OperatorNotSupported)).toExist();
});

it("retrieves CSV when mounted", () => {
  const getCSV = jest.fn();
  shallow(<OperatorInstanceForm {...defaultProps} getCSV={getCSV} />);
  expect(getCSV).toHaveBeenCalledWith(
    defaultProps.cluster,
    defaultProps.namespace,
    defaultProps.csvName,
  );
});

it("retrieves the example values and the target CRD from the given CSV", () => {
  const wrapper = shallow(<OperatorInstanceForm {...defaultProps} />);
  wrapper.setProps({ csv: defaultCSV });
  expect(wrapper.state()).toMatchObject({
    defaultValues: "kind: Foo\napiVersion: v1\n",
    crd: defaultCRD,
  });
});

it("defaults to empty defaultValues if the examples annotation is not found", () => {
  const csv = {
    metadata: {},
    spec: {
      customresourcedefinitions: {
        owned: [defaultCRD],
      },
    },
  } as IClusterServiceVersion;
  const wrapper = shallow(<OperatorInstanceForm {...defaultProps} />);
  wrapper.setProps({ csv });
  expect(wrapper.state()).toMatchObject({
    defaultValues: "",
    crd: defaultCRD,
  });
});

it("renders an error if the CRD is not populated", () => {
  const wrapper = shallow(<OperatorInstanceForm {...defaultProps} />);
  expect(wrapper.find(NotFoundErrorPage)).toExist();
});

it("should submit the form", () => {
  const createResource = jest.fn();
  const wrapper = mount(
    <OperatorInstanceForm {...defaultProps} createResource={createResource} csv={defaultCSV} />,
  );

  const values = "apiVersion: v1\nmetadata:\n  name: foo";
  wrapper.setState({ crd: defaultCRD, defaultValues: values });
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
