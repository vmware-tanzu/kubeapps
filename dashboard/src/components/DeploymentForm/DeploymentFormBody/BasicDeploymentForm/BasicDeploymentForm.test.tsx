// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { mount } from "enzyme";
import { DeploymentEvent, IBasicFormParam } from "shared/types";
import BasicDeploymentForm, { IBasicDeploymentFormProps } from "./BasicDeploymentForm";

jest.useFakeTimers();

const defaultProps = {
  deploymentEvent: "install" as DeploymentEvent,
  handleBasicFormParamChange: jest.fn(() => jest.fn()),
  saveAllChanges: jest.fn(),
  isLoading: false,
  paramsFromComponentState: [],
} as IBasicDeploymentFormProps;

[
  {
    description: "renders a basic deployment with a username",
    params: [
      {
        key: "wordpressUsername",
        currentValue: "user",
        defaultValue: "user",
        deployedValue: "user",
        hasProperties: false,
        title: "Username",
        schema: {
          type: "string",
        },
        type: "string",
      } as IBasicFormParam,
    ],
  },
  {
    description: "renders a basic deployment with a password",
    params: [
      {
        key: "wordpressPassword",
        currentValue: "password",
        defaultValue: "password",
        deployedValue: "password",
        hasProperties: false,
        title: "Password",
        schema: {
          type: "string",
        },
        type: "string",
      } as IBasicFormParam,
    ],
  },
  {
    description: "renders a basic deployment with a email",
    params: [
      {
        key: "wordpressEmail",
        currentValue: "user@example.com",
        defaultValue: "user@example.com",
        deployedValue: "user@example.com",
        hasProperties: false,
        title: "Email",
        schema: {
          type: "string",
        },
        type: "string",
      } as IBasicFormParam,
    ],
  },
  {
    description: "renders a basic deployment with a generic string",
    params: [
      {
        key: "blogName",
        currentValue: "my-blog",
        defaultValue: "my-blog",
        deployedValue: "my-blog",
        hasProperties: false,
        title: "Blog Name",
        schema: {
          type: "string",
        },
        type: "string",
      } as IBasicFormParam,
    ],
  },
  {
    description: "renders a basic deployment with custom configuration",
    params: [
      {
        key: "configuration",
        currentValue: "First line\nSecond line",
        defaultValue: "First line\nSecond line",
        deployedValue: "First line\nSecond line",
        hasProperties: false,
        title: "Configuration",
        schema: {
          type: "object",
        },
        type: "object",
      } as IBasicFormParam,
    ],
  },
  {
    description: "renders a basic deployment with a integer disk size",
    params: [
      {
        key: "size",
        currentValue: 10,
        defaultValue: 10,
        deployedValue: 10,
        hasProperties: false,
        title: "Size",
        schema: {
          type: "integer",
        },
        type: "integer",
      } as IBasicFormParam,
    ],
  },
  {
    description: "renders a basic deployment with a number disk size",
    params: [
      {
        key: "size",
        currentValue: 10.0,
        defaultValue: 10.0,
        deployedValue: 10.0,
        hasProperties: false,
        title: "Size",
        schema: {
          type: "number",
        },
        type: "number",
      } as IBasicFormParam,
    ],
  },
  {
    description: "renders a basic deployment with slider parameters",
    params: [
      {
        key: "size",
        currentValue: 10,
        defaultValue: 10,
        deployedValue: 10,
        hasProperties: false,
        title: "Size",
        schema: {
          type: "integer",
        },
        type: "integer",
        maximum: 100,
        minimum: 1,
      } as IBasicFormParam,
    ],
  },
  {
    description: "renders a basic deployment with username, password, email and a generic string",
    params: [
      {
        key: "wordpressUsername",
        currentValue: "user",
        defaultValue: "user",
        deployedValue: "user",
        hasProperties: false,
        title: "Username",
        schema: {
          type: "string",
        },
        type: "string",
      } as IBasicFormParam,
      {
        key: "wordpressPassword",
        currentValue: "password",
        defaultValue: "password",
        deployedValue: "password",
        hasProperties: false,
        title: "Password",
        schema: {
          type: "string",
        },
        type: "string",
      } as IBasicFormParam,
      {
        key: "wordpressEmail",
        currentValue: "user@example.com",
        defaultValue: "user@example.com",
        deployedValue: "user@example.com",
        hasProperties: false,
        title: "Email",
        schema: {
          type: "string",
        },
        type: "string",
      } as IBasicFormParam,
      {
        key: "blogName",
        currentValue: "my-blog",
        defaultValue: "my-blog",
        deployedValue: "my-blog",
        hasProperties: false,
        title: "Blog Name",
        schema: {
          type: "string",
        },
        type: "string",
      } as IBasicFormParam,
    ],
  },
  {
    description: "renders a basic deployment with a generic boolean",
    params: [
      {
        key: "enableMetrics",
        currentValue: true,
        defaultValue: true,
        deployedValue: true,
        hasProperties: false,
        title: "Metrics",
        schema: {
          type: "boolean",
        },
        type: "boolean",
      } as IBasicFormParam,
    ],
  },
  {
    description: "renders a basic deployment with a generic integer",
    params: [
      {
        key: "replicas",
        currentValue: 10,
        defaultValue: 10,
        deployedValue: 10,
        hasProperties: false,
        title: "Replicas",
        schema: {
          type: "integer",
        },
        type: "integer",
      } as IBasicFormParam,
    ],
  },
  {
    description: "renders an array of strings",
    params: [
      {
        key: "array of strings",
        currentValue: '["element1"]',
        defaultValue: "[element1]",
        deployedValue: "[element1]",
        hasProperties: false,
        minItems: 1,
        title: "string[]",
        schema: {
          type: "array",
          items: {
            type: "string",
          },
        },
        type: "array",
      } as IBasicFormParam,
    ],
  },
  {
    description: "renders an array of numbers",
    params: [
      {
        key: "array of numbers",
        currentValue: "[1]",
        defaultValue: "[1]",
        deployedValue: "[1]",
        hasProperties: false,
        title: "number[]",
        minItems: 1,
        schema: {
          type: "array",
          items: {
            type: "number",
          },
        },
        type: "array",
      } as IBasicFormParam,
    ],
  },
  {
    description: "renders an array of booleans",
    params: [
      {
        key: "array of booleans",
        currentValue: "[true]",
        defaultValue: "[true]",
        deployedValue: "[true]",
        hasProperties: false,
        title: "boolean[]",
        minItems: 1,
        schema: {
          type: "array",
          items: {
            type: "boolean",
          },
        },
        type: "array",
      } as IBasicFormParam,
    ],
  },
  {
    description: "renders an array of objects",
    params: [
      {
        key: "array of objects",
        currentValue: "[{}]",
        defaultValue: "[{}]",
        deployedValue: "[{}]",
        hasProperties: false,
        title: "object[]",
        minItems: 1,
        schema: {
          type: "array",
          items: {
            type: "object",
          },
        },
        type: "array",
      } as IBasicFormParam,
    ],
  },
].forEach(t => {
  it(t.description, () => {
    const onChange = jest.fn();
    const handleBasicFormParamChange = jest.fn(() => onChange);
    const wrapper = mount(
      <BasicDeploymentForm
        {...defaultProps}
        paramsFromComponentState={t.params}
        handleBasicFormParamChange={handleBasicFormParamChange}
      />,
    );

    t.params.forEach((param, i) => {
      const input = wrapper.find(`input#${param.key}`);
      const inputNumText = wrapper.find(`input#${param.key}_text`);
      const inputNumRange = wrapper.find(`input#${param.key}_range`);
      switch (param.type) {
        case "string":
          if (param.key.match("Password")) {
            expect(input.prop("type")).toBe("password");
            break;
          }
          expect(input.prop("type")).toBe("text");
          break;
        case "boolean":
          expect(input.prop("type")).toBe("checkbox");
          break;
        case "number":
          expect(inputNumText.prop("type")).toBe("number");
          expect(inputNumText.prop("step")).toBe(0.5);
          expect(inputNumRange.prop("type")).toBe("range");
          expect(inputNumRange.prop("step")).toBe(0.5);
          break;
        case "integer":
          expect(inputNumText.prop("type")).toBe("number");
          expect(inputNumText.prop("step")).toBe(1);
          expect(inputNumRange.prop("type")).toBe("range");
          expect(inputNumRange.prop("step")).toBe(1);
          break;
        case "array":
          expect(wrapper.find("ArrayParam input").first()).toExist();
          break;
        case "object":
          expect(wrapper.find(`textarea#${param.key}`)).toExist();
          break;
        default:
          break;
      }
      if (["integer", "number"].includes(param.type)) {
        inputNumText.simulate("change", { target: { value: "" } });
      } else if (param.type === "array") {
        wrapper
          .find("ArrayParam input")
          .first()
          .simulate("change", { target: { value: "" } });
      } else if (param.type === "object") {
        wrapper.find(`textarea#${param.key}`).simulate("change", { target: { value: "" } });
      } else {
        input.simulate("change", { target: { value: "" } });
      }
      expect(handleBasicFormParamChange).toHaveBeenCalledWith(param);
      jest.runAllTimers();
      expect(onChange).toHaveBeenCalledTimes(i + 1);
    });
  });
});
