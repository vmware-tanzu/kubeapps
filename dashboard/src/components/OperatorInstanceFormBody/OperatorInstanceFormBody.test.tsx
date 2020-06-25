import { mount, shallow } from "enzyme";
import * as React from "react";
import Modal from "react-modal";
import { Tabs } from "react-tabs";
import OperatorInstanceFormBody from ".";
import itBehavesLike from "../../shared/specs";
import { ConflictError } from "../../shared/types";
import { ErrorSelector } from "../ErrorAlert";
import UnexpectedErrorPage from "../ErrorAlert/UnexpectedErrorAlert";
import { IOperatorInstanceFormProps } from "./OperatorInstanceFormBody";

const defaultProps: IOperatorInstanceFormProps = {
  csvName: "foo",
  isFetching: false,
  namespace: "kubeapps",
  handleDeploy: jest.fn(),
  defaultValues: "",
  errors: {},
};

itBehavesLike("aLoadingComponent", {
  component: OperatorInstanceFormBody,
  props: { ...defaultProps, isFetching: true },
});

it("set default values", () => {
  const wrapper = shallow(<OperatorInstanceFormBody {...defaultProps} />);
  wrapper.setProps({ defaultValues: "foo" });
  expect(wrapper.state()).toMatchObject({
    values: "foo",
  });
});

it("renders an error if there is some error fetching", () => {
  const wrapper = shallow(
    <OperatorInstanceFormBody {...defaultProps} errors={{ fetch: new Error("Boom!") }} />,
  );
  expect(wrapper.find(ErrorSelector)).toExist();
});

it("renders an error if the namespace is _all", () => {
  const wrapper = shallow(<OperatorInstanceFormBody {...defaultProps} namespace="_all" />);
  expect(wrapper.find(UnexpectedErrorPage)).toExist();
});

it("renders an error if the creation failed", () => {
  const wrapper = shallow(<OperatorInstanceFormBody {...defaultProps} />);
  wrapper.setProps({ errors: { create: new ConflictError() } });
  expect(wrapper.find(ErrorSelector)).toExist();
  expect(wrapper.find(Tabs)).toExist();
});

it("restores the default values", async () => {
  const wrapper = mount(<OperatorInstanceFormBody {...defaultProps} />);
  Modal.setAppElement(document.createElement("div"));
  wrapper.setProps({ defaultValues: "foo" });
  wrapper.setState({ values: "not-foo" });
  const restoreButton = wrapper.find("button").filterWhere(b => b.text() === "Restore Defaults");
  restoreButton.simulate("click");

  const restoreConfirmButton = wrapper.find("button").filterWhere(b => b.text() === "Restore");
  restoreConfirmButton.simulate("click");

  const { values } = wrapper.state() as any;
  expect(values).toEqual("foo");
});

it("should submit the form", () => {
  const handleDeploy = jest.fn();
  const wrapper = shallow(
    <OperatorInstanceFormBody {...defaultProps} handleDeploy={handleDeploy} />,
  );

  const values = "apiVersion: v1\nmetadata:\n  name: foo";
  wrapper.setState({ values });
  const form = wrapper.find("form");
  form.simulate("submit", { preventDefault: jest.fn() });

  const resource = {
    apiVersion: "v1",
    metadata: {
      name: "foo",
    },
  };
  expect(handleDeploy).toHaveBeenCalledWith(resource);
});

it("should catch a syntax error in the form", () => {
  const handleDeploy = jest.fn();
  const wrapper = shallow(
    <OperatorInstanceFormBody {...defaultProps} handleDeploy={handleDeploy} />,
  );

  const values = "metadata: invalid!\n  name: foo";
  wrapper.setState({ values });
  const form = wrapper.find("form");
  form.simulate("submit", { preventDefault: jest.fn() });

  expect(
    wrapper
      .find(ErrorSelector)
      .dive()
      .dive()
      .text(),
  ).toContain("Unable to parse the given YAML. Got: bad indentation");
  expect(handleDeploy).not.toHaveBeenCalled();
});

it("should throw an eror if the element doesn't contain an apiVersion", () => {
  const handleDeploy = jest.fn();
  const wrapper = shallow(
    <OperatorInstanceFormBody {...defaultProps} handleDeploy={handleDeploy} />,
  );

  const values = "metadata:\nname: foo";
  wrapper.setState({ values });
  const form = wrapper.find("form");
  form.simulate("submit", { preventDefault: jest.fn() });

  expect(
    wrapper
      .find(ErrorSelector)
      .dive()
      .dive()
      .text(),
  ).toContain("Unable parse the resource. Make sure it contains a valid apiVersion");
  expect(handleDeploy).not.toHaveBeenCalled();
});
