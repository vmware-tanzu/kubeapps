import { shallow } from "enzyme";
import * as React from "react";
import { Tabs } from "react-tabs";
import OperatorInstanceForm from ".";
import itBehavesLike from "../../shared/specs";
import { ConflictError, IClusterServiceVersion } from "../../shared/types";
import { ErrorSelector } from "../ErrorAlert";
import NotFoundErrorPage from "../ErrorAlert/NotFoundErrorAlert";
import { IOperatorInstanceFormProps } from "./OperatorInstanceForm";

const defaultProps: IOperatorInstanceFormProps = {
  csvName: "foo",
  crdName: "foo-cluster",
  isFetching: false,
  namespace: "kubeapps",
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

itBehavesLike("aLoadingComponent", {
  component: OperatorInstanceForm,
  props: { ...defaultProps, isFetching: true },
});

it("retrieves CSV when mounted", () => {
  const getCSV = jest.fn();
  shallow(<OperatorInstanceForm {...defaultProps} getCSV={getCSV} />);
  expect(getCSV).toHaveBeenCalledWith(defaultProps.namespace, defaultProps.csvName);
});

it("retrieves the example values and the target CRD from the given CSV", () => {
  const csv = {
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
  } as IClusterServiceVersion;
  const wrapper = shallow(<OperatorInstanceForm {...defaultProps} />);
  wrapper.setProps({ csv });
  expect(wrapper.state()).toMatchObject({
    defaultValues: "kind: Foo\napiVersion: v1\n",
    crd: defaultCRD,
  });
});

it("renders an error if there is some error fetching", () => {
  const wrapper = shallow(
    <OperatorInstanceForm {...defaultProps} errors={{ fetch: new Error("Boom!") }} />,
  );
  expect(wrapper.find(ErrorSelector)).toExist();
});

it("renders an error if the CRD is not populated", () => {
  const wrapper = shallow(<OperatorInstanceForm {...defaultProps} />);
  expect(wrapper.find(NotFoundErrorPage)).toExist();
});

it("renders an error if the creation failed", () => {
  const wrapper = shallow(<OperatorInstanceForm {...defaultProps} />);
  wrapper.setState({ crd: defaultCRD });
  wrapper.setProps({ errors: { create: new ConflictError() } });
  expect(wrapper.find(ErrorSelector)).toExist();
  expect(wrapper.find(Tabs)).toExist();
});
