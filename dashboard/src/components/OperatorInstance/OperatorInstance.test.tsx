import { shallow } from "enzyme";
import * as React from "react";
import itBehavesLike from "../../shared/specs";
import AppNotes from "../AppView/AppNotes";
import AppValues from "../AppView/AppValues";
import { ErrorSelector } from "../ErrorAlert";
import OperatorInstance, { IOperatorInstanceProps } from "./OperatorInstance";

const defaultProps: IOperatorInstanceProps = {
  isFetching: false,
  namespace: "default",
  csvName: "foo",
  crdName: "foo.kubeapps.com",
  instanceName: "foo-cluster",
  getResource: jest.fn(),
};

itBehavesLike("aLoadingComponent", {
  component: OperatorInstance,
  props: { ...defaultProps, isFetching: true },
});

it("gets a resource when loading the component", () => {
  const getResource = jest.fn();
  shallow(<OperatorInstance {...defaultProps} getResource={getResource} />);
  expect(getResource).toHaveBeenCalledWith(
    defaultProps.namespace,
    defaultProps.csvName,
    defaultProps.crdName,
    defaultProps.instanceName,
  );
});

it("renders an error", () => {
  const wrapper = shallow(<OperatorInstance {...defaultProps} error={new Error("Boom!")} />);
  expect(wrapper.find(ErrorSelector)).toExist();
  expect(wrapper.find(AppNotes)).not.toExist();
});

describe("renders a resource", () => {
  const csv = {
    metadata: { name: "foo" },
    spec: {
      customresourcedefinitions: {
        owned: [{ name: "foo.kubeapps.com", version: "v1alpha1", kind: "Foo" }],
      },
    },
  } as any;
  const resource = { kind: "Foo", spec: { test: true }, status: { alive: true } } as any;
  const props = { ...defaultProps, csv, resource };

  it("renders the resource and CSV info", () => {
    const wrapper = shallow(<OperatorInstance {...props} />);
    expect(wrapper.find(AppNotes)).toExist();
    expect(wrapper.find(AppValues)).toExist();
    expect(wrapper.find(".ChartInfo")).toExist();
    expect(wrapper).toMatchSnapshot();
  });
});
