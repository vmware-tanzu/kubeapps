import { shallow } from "enzyme";
import * as React from "react";
import * as ReactModal from "react-modal";
import itBehavesLike from "../../shared/specs";
import AppNotes from "../AppView/AppNotes";
import AppValues from "../AppView/AppValues";
import ConfirmDialog from "../ConfirmDialog";
import { ErrorSelector } from "../ErrorAlert";
import OperatorInstance, { IOperatorInstanceProps } from "./OperatorInstance";

const defaultProps: IOperatorInstanceProps = {
  isFetching: false,
  namespace: "default",
  csvName: "foo",
  crdName: "foo.kubeapps.com",
  instanceName: "foo-cluster",
  getResource: jest.fn(),
  deleteResource: jest.fn(),
  push: jest.fn(),
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

it("gets a resource again if the namespace changes", () => {
  const getResource = jest.fn();
  const wrapper = shallow(<OperatorInstance {...defaultProps} getResource={getResource} />);
  wrapper.setProps({ namespace: "other" });
  expect(getResource).toHaveBeenCalledTimes(2);
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

  it("deletes the resource", async () => {
    const deleteResource = jest.fn(() => true);
    const push = jest.fn();
    const wrapper = shallow(
      <OperatorInstance {...props} deleteResource={deleteResource} push={push} />,
    );
    wrapper.setState({ crd: csv.spec.customresourcedefinitions.owned[0] });
    ReactModal.setAppElement(document.createElement("div"));
    wrapper.find(".button-danger").simulate("click");

    const dialog = wrapper.find(ConfirmDialog);
    expect(dialog.prop("modalIsOpen")).toEqual(true);
    (dialog.prop("onConfirm") as any)();
    expect(deleteResource).toHaveBeenCalledWith(defaultProps.namespace, "foo", resource);
    // wait async calls
    await new Promise(r => r());
    expect(push).toHaveBeenCalledWith(`/apps/ns/${defaultProps.namespace}`);
  });
});
