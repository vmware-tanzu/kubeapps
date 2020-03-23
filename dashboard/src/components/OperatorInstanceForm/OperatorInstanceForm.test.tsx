import { mount, shallow } from "enzyme";
import * as React from "react";
import * as ReactModal from "react-modal";
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

it("restores the default values", async () => {
  const wrapper = mount(<OperatorInstanceForm {...defaultProps} />);
  ReactModal.setAppElement(document.createElement("div"));
  wrapper.setState({ crd: defaultCRD, values: "bar", defaultValues: "foo" });
  const restoreButton = wrapper.find("button").filterWhere(b => b.text() === "Restore Defaults");
  restoreButton.simulate("click");

  const restoreConfirmButton = wrapper.find("button").filterWhere(b => b.text() === "Restore");
  restoreConfirmButton.simulate("click");

  const { values, defaultValues } = wrapper.state() as any;
  expect(values).toEqual("foo");
  expect(defaultValues).toEqual("foo");
});

it("should submit the form", () => {
  const createResource = jest.fn();
  const wrapper = shallow(
    <OperatorInstanceForm {...defaultProps} createResource={createResource} csv={defaultCSV} />,
  );

  const values = "apiVersion: v1\nmetadata:\n  name: foo";
  wrapper.setState({ crd: defaultCRD, values });
  const form = wrapper.find("form");
  form.simulate("submit", { preventDefault: jest.fn() });

  const resource = {
    apiVersion: "v1",
    metadata: {
      name: "foo",
    },
  };
  expect(createResource).toHaveBeenCalledWith(
    defaultProps.namespace,
    resource.apiVersion,
    defaultCRD.name,
    resource,
  );
});

it("should catch a syntax error in the form", () => {
  const createResource = jest.fn();
  const wrapper = shallow(
    <OperatorInstanceForm {...defaultProps} createResource={createResource} csv={defaultCSV} />,
  );

  const values = "metadata: invalid!\n  name: foo";
  wrapper.setState({ crd: defaultCRD, values });
  const form = wrapper.find("form");
  form.simulate("submit", { preventDefault: jest.fn() });

  expect(
    wrapper
      .find(ErrorSelector)
      .dive()
      .dive()
      .text(),
  ).toContain("Unable to parse the given YAML. Got: bad indentation");
  expect(createResource).not.toHaveBeenCalled();
});

it("should throw an eror if the element doesn't contain an apiVersion", () => {
  const createResource = jest.fn();
  const wrapper = shallow(
    <OperatorInstanceForm {...defaultProps} createResource={createResource} csv={defaultCSV} />,
  );

  const values = "metadata:\nname: foo";
  wrapper.setState({ crd: defaultCRD, values });
  const form = wrapper.find("form");
  form.simulate("submit", { preventDefault: jest.fn() });

  expect(
    wrapper
      .find(ErrorSelector)
      .dive()
      .dive()
      .text(),
  ).toContain("Unable parse the resource. Make sure it contains a valid apiVersion");
  expect(createResource).not.toHaveBeenCalled();
});
