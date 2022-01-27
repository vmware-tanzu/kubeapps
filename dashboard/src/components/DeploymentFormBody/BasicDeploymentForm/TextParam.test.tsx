// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { mount } from "enzyme";
import React from "react";
import { act } from "react-dom/test-utils";
import { IBasicFormParam } from "shared/types";
import TextParam from "./TextParam";

jest.useFakeTimers();

const stringParam = { path: "username", value: "user", type: "string" } as IBasicFormParam;
const stringProps = {
  id: "foo",
  label: "Username",
  param: stringParam,
  handleBasicFormParamChange: jest.fn().mockReturnValue(jest.fn()),
};

it("should render a string parameter with title and description", () => {
  const wrapper = mount(<TextParam {...stringProps} />);
  const input = wrapper.find("input");
  expect(input.prop("value")).toBe(stringProps.param.value);
  expect(wrapper).toMatchSnapshot();
});

it("should set the input type as number", () => {
  const wrapper = mount(<TextParam {...stringProps} inputType={"number"} />);
  const input = wrapper.find("input");
  expect(input.prop("type")).toBe("number");
});

it("should forward the proper value when using a string parameter", () => {
  const handler = jest.fn();
  const handleBasicFormParamChange = jest.fn().mockReturnValue(handler);
  const wrapper = mount(
    <TextParam {...stringProps} handleBasicFormParamChange={handleBasicFormParamChange} />,
  );
  const input = wrapper.find("input");

  const event = { currentTarget: { value: "" } } as React.FormEvent<HTMLInputElement>;
  act(() => {
    (input.prop("onChange") as any)(event);
  });
  wrapper.update();
  jest.runAllTimers();

  expect(handleBasicFormParamChange).toHaveBeenCalledWith({
    path: "username",
    type: "string",
    value: "user",
  });
  expect(handler).toHaveBeenCalledWith(event);
});

it("should set the input value as empty if a string parameter value is not defined", () => {
  const tparam = { path: "username", type: "string" } as IBasicFormParam;
  const tprops = {
    id: "foo",
    name: "username",
    label: "Username",
    param: tparam,
    handleBasicFormParamChange: jest.fn().mockReturnValue(jest.fn()),
  };
  const wrapper = mount(<TextParam {...tprops} />);
  const input = wrapper.find("input");
  expect(input.prop("value")).toBe("");
});

const textAreaParam = {
  path: "configuration",
  value: "First line\nSecond line",
  type: "string",
} as IBasicFormParam;
const textAreaProps = {
  id: "bar",
  label: "Configuration",
  param: textAreaParam,
  handleBasicFormParamChange: jest.fn().mockReturnValue(jest.fn()),
  inputType: "textarea",
};

it("should render a textArea parameter with title and description", () => {
  const wrapper = mount(<TextParam {...textAreaProps} />);
  const input = wrapper.find("textarea");
  expect(input.prop("value")).toBe(textAreaProps.param.value);
  expect(wrapper).toMatchSnapshot();
});

it("should forward the proper value when using a textArea parameter", () => {
  const handler = jest.fn();
  const handleBasicFormParamChange = jest.fn().mockReturnValue(handler);
  const wrapper = mount(
    <TextParam {...textAreaProps} handleBasicFormParamChange={handleBasicFormParamChange} />,
  );
  const input = wrapper.find("textarea");

  const event = { currentTarget: { value: "" } } as React.FormEvent<HTMLInputElement>;
  act(() => {
    (input.prop("onChange") as any)(event);
  });
  wrapper.update();
  jest.runAllTimers();

  expect(handleBasicFormParamChange).toHaveBeenCalledWith({
    path: "configuration",
    type: "string",
    value: "First line\nSecond line",
  });
  expect(handler).toHaveBeenCalledWith(event);
});

it("should set the input value as empty if a textArea param value is not defined", () => {
  const tparam = { path: "configuration", type: "string" } as IBasicFormParam;
  const tprops = {
    id: "foo",
    name: "configuration",
    label: "Configuration",
    param: tparam,
    handleBasicFormParamChange: jest.fn().mockReturnValue(jest.fn()),
    inputType: "textarea",
  };
  const wrapper = mount(<TextParam {...tprops} />);
  const input = wrapper.find("textarea");
  expect(input.prop("value")).toBe("");
});

it("should render a string parameter as select with option tags", () => {
  const tparam = {
    path: "databaseType",
    value: "postgresql",
    type: "string",
    enum: ["mariadb", "postgresql"],
  } as IBasicFormParam;
  const tprops = {
    id: "foo",
    name: "databaseType",
    label: "databaseType",
    param: tparam,
    handleBasicFormParamChange: jest.fn().mockReturnValue(jest.fn()),
  };
  const wrapper = mount(<TextParam {...tprops} />);
  const input = wrapper.find("select");

  expect(wrapper.find("select").prop("value")).toBe(tparam.value);
  if (tparam.enum != null) {
    const options = input.find("option");
    expect(options.length).toBe(tparam.enum.length);

    for (let i = 0; i < tparam.enum.length; i++) {
      const option = options.at(i);
      expect(option.text()).toBe(tparam.enum[i]);
    }
  }
});

it("should forward the proper value when using a select", () => {
  const tparam = {
    path: "databaseType",
    value: "postgresql",
    type: "string",
    enum: ["mariadb", "postgresql"],
  } as IBasicFormParam;
  const tprops = {
    id: "foo",
    name: "databaseType",
    label: "databaseType",
    param: tparam,
  };
  const handler = jest.fn();
  const handleBasicFormParamChange = jest.fn().mockReturnValue(handler);
  const wrapper = mount(
    <TextParam {...tprops} handleBasicFormParamChange={handleBasicFormParamChange} />,
  );
  const input = wrapper.find("select");

  const event = { currentTarget: {} } as React.FormEvent<HTMLSelectElement>;
  act(() => {
    (input.prop("onChange") as any)(event);
  });

  expect(handleBasicFormParamChange.mock.calls[0][0]).toEqual({
    path: "databaseType",
    type: "string",
    value: "postgresql",
    enum: ["mariadb", "postgresql"],
  });
  expect(handler.mock.calls[0][0]).toMatchObject(event);
});

it("a change in the param property should update the current value", () => {
  const wrapper = mount(<TextParam {...stringProps} param={{ ...stringParam, value: "" }} />);
  const input = wrapper.find("input");
  expect(input.prop("value")).toBe("");

  wrapper.setProps({
    param: {
      ...stringParam,
      value: "foo",
    },
  });
  wrapper.update();
  expect(wrapper.find("input").prop("value")).toBe("foo");
});

it("a change in a number param property should update the current value", () => {
  const numberParam = { path: "replicas", value: 0, type: "number" } as IBasicFormParam;
  const wrapper = mount(<TextParam {...stringProps} param={numberParam} />);
  const input = wrapper.find("input");
  expect(input.prop("value")).toBe(0);

  wrapper.setProps({
    param: {
      ...numberParam,
      value: 1,
    },
  });
  wrapper.update();
  expect(wrapper.find("input").prop("value")).toBe(1);
});
