import * as React from "react";

import { mount } from "enzyme";
import TextParam from "./TextParam";

const param = { path: "username", value: "user", type: "string" };
const defaultProps = {
  id: "foo",
  label: "Username",
  param,
  handleBasicFormParamChange: jest.fn(() => jest.fn()),
};

it("should render a text param with title and description", () => {
  const wrapper = mount(<TextParam {...defaultProps} />);
  const input = wrapper.find("input");
  expect(input.prop("value")).toBe(defaultProps.param.value);
  expect(wrapper).toMatchSnapshot();
});

it("should set the input type as number", () => {
  const wrapper = mount(<TextParam {...defaultProps} inputType={"number"} />);
  const input = wrapper.find("input");
  expect(input.prop("type")).toBe("number");
});

it("should forward the proper value", () => {
  const handler = jest.fn();
  const handleBasicFormParamChange = jest.fn(() => handler);
  const wrapper = mount(
    <TextParam {...defaultProps} handleBasicFormParamChange={handleBasicFormParamChange} />,
  );
  const input = wrapper.find("input");

  const event = { currentTarget: {} } as React.FormEvent<HTMLInputElement>;
  (input.prop("onChange") as any)(event);

  expect(handleBasicFormParamChange.mock.calls[0][0]).toEqual({
    path: "username",
    type: "string",
    value: "user",
  });
  expect(handler.mock.calls[0][0]).toMatchObject(event);
});

it("should set the input value as empty if the param value is not defined", () => {
  const tparam = { path: "username", type: "string" };
  const tprops = {
    id: "foo",
    name: "username",
    label: "Username",
    param: tparam,
    handleBasicFormParamChange: jest.fn(() => jest.fn()),
  };
  const wrapper = mount(<TextParam {...tprops} />);
  const input = wrapper.find("input");
  expect(input.prop("value")).toBe("");
});
