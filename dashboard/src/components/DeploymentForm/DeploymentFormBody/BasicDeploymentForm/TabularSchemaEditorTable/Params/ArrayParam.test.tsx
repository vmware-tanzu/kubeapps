// Copyright 2022-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import { act } from "@testing-library/react";
import { mount } from "enzyme";
import { IBasicFormParam } from "shared/types";
import ArrayParam, { IArrayParamProps } from "./ArrayParam";

jest.useFakeTimers();

const arrayParams = [
  {
    description: "array of numbers",
    props: {
      handleBasicFormParamChange: jest.fn(),
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
      } as IBasicFormParam,
    } as IArrayParamProps,
  },
  {
    description: "array of strings",
    props: {
      handleBasicFormParamChange: jest.fn(),
      id: "array-strings",
      label: "label",
      type: "string",
      param: {
        key: "array of numbers",
        currentValue: '["element1"]',
        defaultValue: '["element1"]',
        deployedValue: '["element1"]',
        hasProperties: false,
        title: "string[]",
        schema: {
          type: "array",
          items: {
            type: "string",
          },
        },
        type: "array",
      } as IBasicFormParam,
    } as IArrayParamProps,
  },
  {
    description: "array of booleans",
    props: {
      handleBasicFormParamChange: jest.fn(),
      id: "array-boolean",
      label: "label",
      type: "boolean",
      param: {
        key: "array of booleans",
        currentValue: "[true]",
        defaultValue: "[true]",
        deployedValue: "[true]",
        hasProperties: false,
        title: "boolean[]",
        schema: {
          type: "array",
          items: {
            type: "boolean",
          },
        },
        type: "array",
      } as IBasicFormParam,
    } as IArrayParamProps,
  },
  {
    description: "array of objects",
    props: {
      handleBasicFormParamChange: jest.fn(),
      id: "array-object",
      label: "label",
      type: "object",
      param: {
        key: "array of objects",
        currentValue: "[{}]",
        defaultValue: "[{}]",
        deployedValue: "[{}]",
        hasProperties: false,
        title: "object[]",
        schema: {
          type: "array",
          items: {
            type: "object",
          },
        },
        type: "array",
      } as IBasicFormParam,
    } as IArrayParamProps,
  },
  {
    description: "array of arrays",
    props: {
      handleBasicFormParamChange: jest.fn(),
      id: "array-object",
      label: "label",
      type: "object",
      param: {
        key: "array of objects",
        currentValue: "[[]]",
        defaultValue: "[[]]",
        deployedValue: "[[]]",
        hasProperties: false,
        title: "object[]",
        schema: {
          type: "array",
          items: {
            type: "object",
          },
        },
        type: "array",
      } as IBasicFormParam,
    } as IArrayParamProps,
  },
];

arrayParams.forEach(t => {
  it(`renders an ${t.description}`, () => {
    const onChange = jest.fn();
    const handleBasicFormParamChange = jest.fn(() => onChange);
    const wrapper = mount(
      <ArrayParam {...t.props} handleBasicFormParamChange={handleBasicFormParamChange} />,
    );

    // Add a new element to create the input field
    act(() => {
      wrapper
        .find(CdsButton)
        .filterWhere(b => b.text() === "Add")
        .simulate("click");
    });
    wrapper.update();

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
        input.simulate("change", { target: { value: "" } });
        break;
      case "boolean":
        expect(input.prop("type")).toBe("checkbox");
        input.simulate("change", { target: { value: "" } });
        break;
      case "number":
        expect(inputNumText.prop("type")).toBe("number");
        expect(inputNumText.prop("step")).toBe(0.5);
        expect(inputNumRange.prop("type")).toBe("range");
        expect(inputNumRange.prop("step")).toBe(0.5);
        inputNumText.simulate("change", { target: { value: "" } });
        break;
      case "integer":
        expect(inputNumText.prop("type")).toBe("number");
        expect(inputNumText.prop("step")).toBe(1);
        expect(inputNumRange.prop("type")).toBe("range");
        expect(inputNumRange.prop("step")).toBe(1);
        inputNumText.simulate("change", { target: { value: "" } });
        break;
      case "object":
        expect(input).toExist();
        input.simulate("change", { target: { value: "" } });
        break;
      case "array":
        expect(input).toExist();
        input.simulate("change", { target: { value: "" } });
        break;
      default:
        break;
    }
    expect(handleBasicFormParamChange).toHaveBeenCalledWith(t.props.param);
    jest.runAllTimers();
    expect(onChange).toHaveBeenCalledTimes(1);
  });
});

arrayParams.forEach(t => {
  it(`uses the minItems and maxItems in an ${t.description}`, () => {
    const wrapper = mount(
      <ArrayParam {...t.props} param={{ ...t.props.param, minItems: 1, maxItems: 1 }} />,
    );

    const inputNumText = wrapper.find(`input#${t.props.id}-0_text`);
    const inputNumRange = wrapper.find(`input#${t.props.id}-0_range`);
    const input = wrapper.find("input");
    const arrayType = t.props.param.schema.items.type;

    switch (arrayType) {
      case "string":
        expect(input.length).toBe(1);
        break;
      case "boolean":
        expect(input.length).toBe(1);
        break;
      case "number":
        expect(inputNumText.length).toBe(1);
        expect(inputNumRange.length).toBe(1);
        break;
      case "integer":
        expect(inputNumText.length).toBe(1);
        expect(inputNumRange.length).toBe(1);
        break;
      case "object":
        expect(input.length).toBe(1);
        break;
      case "array":
        expect(input.length).toBe(1);
        break;
      default:
        break;
    }

    expect(
      wrapper
        .find(CdsButton)
        .filterWhere(b => b.text() === "Add")
        .prop("disabled"),
    ).toBe(true);
  });
});

arrayParams.forEach(t => {
  it(`uses the required property in an ${t.description}`, () => {
    const wrapper = mount(
      <ArrayParam {...t.props} param={{ ...t.props.param, isRequired: true }} />,
    );

    // Add a new element to create the input field
    act(() => {
      wrapper
        .find(CdsButton)
        .filterWhere(b => b.text() === "Add")
        .simulate("click");
    });
    wrapper.update();

    const inputNumText = wrapper.find(`input#${t.props.id}-0_text`);
    const inputNumRange = wrapper.find(`input#${t.props.id}-0_range`);
    const input = wrapper.find("input");
    const arrayType = t.props.param.schema.items.type;

    switch (arrayType) {
      case "string":
        expect(input.prop("required")).toBe(true);
        break;
      case "boolean":
        expect(input.prop("required")).toBe(true);
        break;
      case "number":
        expect(inputNumText.prop("required")).toBe(true);
        expect(inputNumRange.prop("required")).toBe(true);
        break;
      case "integer":
        expect(inputNumText.prop("required")).toBe(true);
        expect(inputNumRange.prop("required")).toBe(true);
        break;
      case "object":
        expect(input.prop("required")).toBe(true);
        break;
      case "array":
        expect(input.prop("required")).toBe(true);
        break;
      default:
        break;
    }
  });
});

arrayParams.forEach(t => {
  it(`uses the readOnly property in an ${t.description}`, () => {
    const wrapper = mount(<ArrayParam {...t.props} param={{ ...t.props.param, readOnly: true }} />);

    // Add a new element to create the input field
    act(() => {
      wrapper
        .find(CdsButton)
        .filterWhere(b => b.text() === "Add")
        .simulate("click");
    });
    wrapper.update();

    const inputNumText = wrapper.find(`input#${t.props.id}-0_text`);
    const inputNumRange = wrapper.find(`input#${t.props.id}-0_range`);
    const input = wrapper.find("input");
    const arrayType = t.props.param.schema.items.type;

    switch (arrayType) {
      case "string":
        expect(input.prop("disabled")).toBe(true);
        break;
      case "boolean":
        expect(input.prop("disabled")).toBe(true);
        break;
      case "number":
        expect(inputNumText.prop("disabled")).toBe(true);
        expect(inputNumRange.prop("disabled")).toBe(true);
        break;
      case "integer":
        expect(inputNumText.prop("disabled")).toBe(true);
        expect(inputNumRange.prop("disabled")).toBe(true);
        break;
      case "object":
        expect(input.prop("disabled")).toBe(true);
        break;
      case "array":
        expect(input.prop("disabled")).toBe(true);
        break;
      default:
        break;
    }
  });
});

arrayParams
  .filter(t => ["integer", "number"].includes(t.props.type))
  .forEach(t => {
    it(`uses the max/min property in an ${t.description}`, () => {
      const wrapper = mount(
        <ArrayParam
          {...t.props}
          param={{
            ...t.props.param,
            minimum: 7,
            exclusiveMinimum: 5,
            maximum: 55,
            exclusiveMaximum: 50,
          }}
        />,
      );

      // Add a new element to create the input field
      act(() => {
        wrapper
          .find(CdsButton)
          .filterWhere(b => b.text() === "Add")
          .simulate("click");
      });
      wrapper.update();

      const inputNumText = wrapper.find(`input#${t.props.id}-0_text`);
      const inputNumRange = wrapper.find(`input#${t.props.id}-0_range`);

      expect(inputNumText.prop("max")).toBe(50);
      expect(inputNumRange.prop("max")).toBe(50);

      expect(inputNumText.prop("min")).toBe(5);
      expect(inputNumRange.prop("min")).toBe(5);
    });
  });

arrayParams
  .filter(t => ["integer", "number"].includes(t.props.type))
  .forEach(t => {
    it(`uses the multipleOf property in an ${t.description}`, () => {
      const wrapper = mount(
        <ArrayParam
          {...t.props}
          param={{
            ...t.props.param,
            multipleOf: 5,
          }}
        />,
      );

      // Add a new element to create the input field
      act(() => {
        wrapper
          .find(CdsButton)
          .filterWhere(b => b.text() === "Add")
          .simulate("click");
      });
      wrapper.update();

      const inputNumText = wrapper.find(`input#${t.props.id}-0_text`);
      const inputNumRange = wrapper.find(`input#${t.props.id}-0_range`);

      expect(inputNumText.prop("step")).toBe(5);
      expect(inputNumRange.prop("step")).toBe(5);

      expect(inputNumText.prop("step")).toBe(5);
      expect(inputNumRange.prop("step")).toBe(5);
    });
  });
