// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { mount } from "enzyme";
import ArrayParam, { IArrayParamProps } from "./ArrayParam";

jest.useFakeTimers();

[
  {
    description: "renders an array of numbers",
    props: {
      id: "array-numbers",
      label: "label",
      type: "number",
      param: {
        key: "array of numbers",
        currentValue: "[1]",
        defaultValue: "[1]",
        deployedValue: "[1]",
        hasProperties: false,
        title: "number[]",
        schema: {
          type: "array",
          items: {
            type: "number",
          },
        },
        type: "array",
      },
    } as IArrayParamProps,
  },
  {
    description: "renders an array of strings",
    props: {
      id: "array-strings",
      label: "label",
      type: "string",
      param: {
        key: "array of numbers",
        currentValue: '["element1"]',
        defaultValue: "[element1]",
        deployedValue: "[element1]",
        hasProperties: false,
        title: "string[]",
        schema: {
          type: "array",
          items: {
            type: "string",
          },
        },
        type: "array",
      },
    } as IArrayParamProps,
  },
  {
    description: "renders an array of booleans",
    props: {
      id: "array-boolean",
      label: "label",
      type: "boolean",
      param: {
        key: "array of booleans",
        currentValue: "[1]",
        defaultValue: "[1]",
        deployedValue: "[1]",
        hasProperties: false,
        title: "boolean[]",
        schema: {
          type: "array",
          items: {
            type: "boolean",
          },
        },
        type: "array",
      },
    } as IArrayParamProps,
  },
  {
    description: "renders an array of objects",
    props: {
      id: "array-object",
      label: "label",
      type: "object",
      param: {
        key: "array of objects",
        currentValue: "[1]",
        defaultValue: "[1]",
        deployedValue: "[1]",
        hasProperties: false,
        title: "object[]",
        schema: {
          type: "array",
          items: {
            type: "object",
          },
        },
        type: "array",
      },
    } as IArrayParamProps,
  },
].forEach(t => {
  it(t.description, () => {
    const onChange = jest.fn();
    const handleBasicFormParamChange = jest.fn(() => onChange);
    const wrapper = mount(
      <ArrayParam {...t.props} handleBasicFormParamChange={handleBasicFormParamChange} />,
    );
    const inputNumText = wrapper.find(`input#${t.props.id}-0_text`);
    const inputNumRange = wrapper.find(`input#${t.props.id}-0_range`);
    const input = wrapper.find("input");
    const arrayType = t.props.param.schema.items.type;

    switch (arrayType) {
      case "string":
        if (t.props.param.key.match("Password")) {
          expect(input.prop("type")).toBe("password");
          break;
        }
        expect(input).toExist();
        break;
      case "boolean":
        expect(input.prop("type")).toBe("checkbox");
        break;
      case "number":
        expect(inputNumText.prop("type")).toBe("number");
        expect(inputNumText.prop("step")).toBe(0.1);
        expect(inputNumRange.prop("type")).toBe("range");
        expect(inputNumRange.prop("step")).toBe(0.1);
        break;
      case "integer":
        expect(inputNumText.prop("type")).toBe("number");
        expect(inputNumText.prop("step")).toBe(1);
        expect(inputNumRange.prop("type")).toBe("range");
        expect(inputNumRange.prop("step")).toBe(1);
        break;
      case "object":
        expect(input).toExist();
        break;
      default:
        break;
    }
    if (["integer", "number"].includes(arrayType)) {
      inputNumText.simulate("change", { target: { value: "" } });
    } else {
      input.simulate("change", { target: { value: "" } });
    }
    expect(handleBasicFormParamChange).toHaveBeenCalledWith(t.props.param);
    jest.runAllTimers();
    expect(onChange).toHaveBeenCalledTimes(1);
  });
});
