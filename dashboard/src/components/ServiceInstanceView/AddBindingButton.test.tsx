import { shallow } from "enzyme";
import context from "jest-plugin-context";
import * as React from "react";

import { JSONSchema6 } from "json-schema";
import { ErrorSelector } from "../../components/ErrorAlert";
import SchemaForm from "../../components/SchemaForm";
import AddBindingButton from "./AddBindingButton";

const defaultName = "my-class";
const defaultNS = "default";
const defaultProps = {
  disabled: false,
  instanceRefName: defaultName,
  namespace: defaultNS,
  addBinding: jest.fn(),
  onAddBinding: jest.fn(),
};

it("shows a button", () => {
  const wrapper = shallow(<AddBindingButton {...defaultProps} />);
  wrapper.setState({ isProvisioning: true });
  expect(wrapper.find(".button")).toExist();
});

it("disables the button", () => {
  const wrapper = shallow(<AddBindingButton {...defaultProps} disabled={true} />);
  wrapper.setState({ isProvisioning: true });
  expect(
    wrapper
      .find(".button")
      .filterWhere(e => e.text() === "Add Binding")
      .prop("disabled"),
  ).toBe(true);
});

context("when the modal is open", () => {
  it("shows an error if is present", () => {
    const wrapper = shallow(<AddBindingButton {...defaultProps} error={new Error()} />);
    wrapper.setState({ modalIsOpen: true });

    expect(wrapper.find(ErrorSelector)).toExist();
    expect(wrapper).toMatchSnapshot();
  });

  it("should display the name form", () => {
    const wrapper = shallow(<AddBindingButton {...defaultProps} />);
    wrapper.setState({ modalIsOpen: true, displayNameForm: true });

    const form = wrapper.find(SchemaForm);
    expect(form).toExist();
    expect((form.prop("schema") as any).properties).toHaveProperty("Name");
  });

  it("should display the default schema if it's not in the plan spec", () => {
    const wrapper = shallow(<AddBindingButton {...defaultProps} bindingSchema={undefined} />);
    wrapper.setState({ modalIsOpen: true, displayNameForm: false });

    const form = wrapper.find(SchemaForm);
    expect(form).toExist();
    const schema = form.prop("schema") as any;
    expect(schema).toMatchObject({
      properties: {
        kubeappsRawParameters: {
          title: "Parameters",
          type: "object",
        },
      },
      type: "object",
    });
  });

  it("should display the spec schema", () => {
    const newSchema = {
      properties: {
        kubeappsRawParameters: {
          title: "Foo",
          type: "string",
        },
      },
      type: "object",
    } as JSONSchema6;
    const wrapper = shallow(<AddBindingButton {...defaultProps} bindingSchema={newSchema} />);
    wrapper.setState({ modalIsOpen: true, displayNameForm: false });

    const form = wrapper.find(SchemaForm);
    expect(form).toExist();
    const schema = form.prop("schema") as any;
    expect(schema).toEqual(newSchema);
  });
});
