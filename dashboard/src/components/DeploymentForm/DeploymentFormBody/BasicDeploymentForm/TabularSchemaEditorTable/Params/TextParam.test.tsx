// Copyright 2019-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { act } from "@testing-library/react";
import { mount } from "enzyme";
import { IBasicFormParam } from "shared/types";
import TextParam, { ITextParamProps } from "./TextParam";

jest.useFakeTimers();

describe("param rendered as a input type text", () => {
  const stringParam = {
    currentValue: "foo",
    defaultValue: "foo",
    deployedValue: "foo",
    hasProperties: false,
    title: "Username",
    schema: {
      type: "string",
    },
    type: "string",
    key: "username",
  } as IBasicFormParam;
  const stringProps = {
    id: "foo-string",
    label: "Username",
    inputType: "text",
    param: stringParam,
    handleBasicFormParamChange: jest.fn().mockReturnValue(jest.fn()),
  } as ITextParamProps;

  it("should render a string parameter with title and description", () => {
    const wrapper = mount(<TextParam {...stringProps} />);
    const input = wrapper.find("input");
    expect(input.prop("value")).toBe(stringProps.param.currentValue);
  });

  it("should forward the proper value when using a string parameter", () => {
    const handler = jest.fn();
    const handleBasicFormParamChange = jest.fn().mockReturnValue(handler);
    const wrapper = mount(
      <TextParam {...stringProps} handleBasicFormParamChange={handleBasicFormParamChange} />,
    );
    const input = wrapper.find("input");

    const event = {
      currentTarget: { value: "", reportValidity: jest.fn() },
    } as unknown as React.FormEvent<HTMLInputElement>;
    act(() => {
      (input.prop("onChange") as any)(event);
    });
    wrapper.update();
    jest.runAllTimers();

    expect(handleBasicFormParamChange).toHaveBeenCalledWith(stringProps.param);
    expect(handler).toHaveBeenCalledWith({
      ...event,
      currentTarget: { ...event.currentTarget, reportValidity: undefined },
    });
  });

  it("should set the input value as empty if a string parameter value is not defined", () => {
    const wrapper = mount(
      <TextParam
        {...{ ...stringProps, param: { ...stringProps.param, currentValue: undefined } }}
      />,
    );
    const input = wrapper.find("input");
    expect(input.prop("value")).toBe("");
  });

  it("a change in the param property should update the current value", () => {
    const wrapper = mount(
      <TextParam {...stringProps} param={{ ...stringParam, currentValue: "" }} />,
    );
    const input = wrapper.find("input");
    expect(input.prop("value")).toBe("");

    const event = {
      currentTarget: { value: "foo", reportValidity: jest.fn() },
    } as unknown as React.FormEvent<HTMLInputElement>;
    act(() => {
      (input.prop("onChange") as any)(event);
    });
    wrapper.update();
    expect(wrapper.find("input").prop("value")).toBe("foo");
  });
});

describe("param rendered as a input type textarea", () => {
  const textAreaParam = {
    type: "string",
    schema: {
      type: "string",
    },
    key: "configuration",
    title: "Configuration",
    currentValue: "First line\nSecond line",
    defaultValue: "First line\nSecond line",
    deployedValue: "First line\nSecond line",
    hasProperties: false,
  } as IBasicFormParam;
  const textAreaProps = {
    id: "foo-textarea",
    label: "Configuration",
    param: textAreaParam,
    handleBasicFormParamChange: jest.fn().mockReturnValue(jest.fn()),
    inputType: "textarea",
  } as ITextParamProps;

  it("should render a textArea parameter with title and description", () => {
    const wrapper = mount(<TextParam {...textAreaProps} />);
    const input = wrapper.find("textarea");
    expect(input.prop("value")).toBe(textAreaProps.param.currentValue);
  });

  it("should forward the proper value when using a textArea parameter", () => {
    const handler = jest.fn();
    const handleBasicFormParamChange = jest.fn().mockReturnValue(handler);
    const wrapper = mount(
      <TextParam {...textAreaProps} handleBasicFormParamChange={handleBasicFormParamChange} />,
    );
    const input = wrapper.find("textarea");

    const event = {
      currentTarget: { value: "", reportValidity: jest.fn() },
    } as unknown as React.FormEvent<HTMLInputElement>;
    act(() => {
      (input.prop("onChange") as any)(event);
    });
    wrapper.update();
    jest.runAllTimers();

    expect(handleBasicFormParamChange).toHaveBeenCalledWith(textAreaParam);
    expect(handler).toHaveBeenCalledWith({
      ...event,
      currentTarget: { ...event.currentTarget, reportValidity: undefined },
    });
  });

  it("should set the input value as empty if a textArea param value is not defined", () => {
    const wrapper = mount(
      <TextParam
        {...{ ...textAreaProps, param: { ...textAreaProps.param, currentValue: undefined } }}
      />,
    );
    const input = wrapper.find("textarea");
    expect(input.prop("value")).toBe("");
  });
});

describe("param rendered as a select", () => {
  const enumParam = {
    type: "string",
    schema: {
      type: "string",
      enum: ["mariadb", "postgresql"],
    },
    enum: ["mariadb", "postgresql"],
    key: "databaseType",
    title: "Database Type",
    currentValue: "postgresql",
    defaultValue: "postgresql",
    deployedValue: "postgresql",
    hasProperties: false,
    isRequired: false,
  } as IBasicFormParam;
  const enumProps = {
    id: "foo-enum",
    name: "databaseType",
    label: "Database Type",
    param: enumParam,
    handleBasicFormParamChange: jest.fn().mockReturnValue(jest.fn()),
  } as ITextParamProps;

  it("should render a string parameter as select with option tags", () => {
    const wrapper = mount(<TextParam {...enumProps} />);
    const input = wrapper.find("select");

    expect(wrapper.find("select").prop("value")).toBe(enumParam.currentValue);
    if (enumParam.enum != null) {
      const options = input.find("option");
      expect(options.length).toBe(enumParam.enum.length + 1); //empty default option

      expect(options.at(0).text()).toBe("");
      for (let i = 1; i < enumParam.enum.length; i++) {
        const option = options.at(i);
        expect(option.text()).toBe(enumParam.enum[i - 1]);
      }
    }
  });

  it("should forward the proper value when using a select", () => {
    const handler = jest.fn();
    const handleBasicFormParamChange = jest.fn().mockReturnValue(handler);
    const wrapper = mount(
      <TextParam {...enumProps} handleBasicFormParamChange={handleBasicFormParamChange} />,
    );
    expect(wrapper.find("select").prop("value")).toBe(enumParam.currentValue);

    const event = {
      currentTarget: { value: "mariadb", type: undefined, reportValidity: jest.fn() },
    } as unknown as React.FormEvent<HTMLSelectElement>;
    act(() => {
      (wrapper.find("select").prop("onChange") as any)(event);
    });
    wrapper.update();
    jest.runAllTimers();

    expect(wrapper.find("select").prop("value")).toBe(event.currentTarget.value);
    expect(handleBasicFormParamChange).toHaveBeenCalledWith(enumProps.param);
    expect(handler).toHaveBeenCalledWith({
      ...event,
      currentTarget: { ...event.currentTarget, reportValidity: undefined },
    });
  });
});
