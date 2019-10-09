import * as React from "react";

import { mount } from "enzyme";
import TextParam from "./TextParam";

const param = { path: "username", value: "user", type: "string" };
const defaultProps = {
  id: "foo",
  name: "username",
  label: "Username",
  param,
  handleBasicFormParamChange: jest.fn(),
};

it("should render a text param with title and description", () => {
  const wrapper = mount(<TextParam {...defaultProps} />);
  const input = wrapper.find("input");
  expect(input.prop("defaultValue")).toBe(defaultProps.param.value);
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

  expect(handleBasicFormParamChange.mock.calls[0][0]).toBe("username");
  expect(handler.mock.calls[0][0]).toMatchObject(event);
});
