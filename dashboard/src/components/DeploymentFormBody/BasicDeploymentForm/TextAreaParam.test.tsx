import * as React from "react";

import { mount } from "enzyme";
import { IBasicFormParam } from "shared/types";
import TextAreaParam from "./TextAreaParam";

const param = {
  path: "configuration",
  value: "First line\n" + "Second line",
  type: "string",
} as IBasicFormParam;
const defaultProps = {
  id: "foo",
  label: "Configuration",
  param,
  handleBasicFormParamChange: jest.fn(() => jest.fn()),
};

it("should render a textArea param with title and description", () => {
  const wrapper = mount(<TextAreaParam {...defaultProps} />);
  const input = wrapper.find("textarea");
  expect(input.prop("value")).toBe(defaultProps.param.value);
  expect(wrapper).toMatchSnapshot();
});

it("should forward the proper value", () => {
  const handler = jest.fn();
  const handleBasicFormParamChange = jest.fn(() => handler);
  const wrapper = mount(
    <TextAreaParam {...defaultProps} handleBasicFormParamChange={handleBasicFormParamChange} />,
  );
  const input = wrapper.find("textarea");

  const event = { currentTarget: {} } as React.FormEvent<HTMLInputElement>;
  (input.prop("onChange") as any)(event);

  expect(handleBasicFormParamChange.mock.calls[0][0]).toEqual({
    path: "configuration",
    type: "string",
    value: "First line\n" + "Second line",
  });
  expect(handler.mock.calls[0][0]).toMatchObject(event);
});

it("should set the input value as empty if the param value is not defined", () => {
  const tparam = { path: "configuration" } as IBasicFormParam;
  const tprops = {
    id: "foo",
    name: "configuration",
    label: "Configuration",
    param: tparam,
    handleBasicFormParamChange: jest.fn(() => jest.fn()),
  };
  const wrapper = mount(<TextAreaParam {...tprops} />);
  const input = wrapper.find("textarea");
  expect(input.prop("value")).toBe("");
});
