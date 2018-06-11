import { shallow } from "enzyme";
import { JSONSchema6 } from "json-schema";
import * as React from "react";
import { FieldProps } from "react-jsonschema-form";
import ObjectField from "react-jsonschema-form/lib/components/fields/ObjectField";

import AceEditor from "react-ace";
import CustomObjectField from "./CustomObjectField";

it("should render the upstream ObjectField if schema has properties", () => {
  const schema: JSONSchema6 = {
    properties: {
      firstName: {
        type: "string",
      },
    },
    type: "object",
  };
  const props = {
    onChange: (value: any) => {
      jest.fn();
    },
    schema,
  } as FieldProps;

  const wrapper = shallow(<CustomObjectField {...props} />);
  const field = wrapper.find(ObjectField);
  expect(field.exists()).toBe(true);
  expect(field.props()).toMatchObject(props);

  const editor = wrapper.find(AceEditor);
  expect(editor.exists()).toBe(false);
});

it("should render the raw JSON editor if schema doesn't have properties", () => {
  const schema: JSONSchema6 = {
    additionalProperties: {
      type: "string",
    },
    description: "Tags to be applied to new resources, specified as key/value pairs.",
    type: "object",
  };
  const props = {
    onChange: (value: any) => {
      jest.fn();
    },
    schema,
  } as FieldProps;
  const wrapper = shallow(<CustomObjectField {...props} />);
  const field = wrapper.find(ObjectField);
  expect(field.exists()).toBe(false);

  const editor = wrapper.find(AceEditor);
  expect(editor.exists()).toBe(true);
});

it("should populate the raw JSON editor with a default if defined in schema", () => {
  const defaultValue = JSON.stringify({ winter: "is coming" });
  const schema: JSONSchema6 = {
    additionalProperties: {
      type: "string",
    },
    default: defaultValue,
    description: "Tags to be applied to new resources, specified as key/value pairs.",
    type: "object",
  };
  const props = {
    onChange: (value: any) => {
      jest.fn();
    },
    schema,
  } as FieldProps;
  const wrapper = shallow(<CustomObjectField {...props} />);
  expect(wrapper.state("rawValue")).toBe(defaultValue);
});

it("should trigger the onChange handler when changed", () => {
  const schema: JSONSchema6 = {
    additionalProperties: {
      type: "string",
    },
    description: "Tags to be applied to new resources, specified as key/value pairs.",
    type: "object",
  };
  const onChange = jest.fn();
  const props = {
    onChange: (value: any) => {
      onChange(value);
    },
    schema,
  } as FieldProps;
  const wrapper = shallow(<CustomObjectField {...props} />);

  const editor = wrapper.find(AceEditor);
  editor.simulate("change", '{"test":true}');
  expect(wrapper.state("rawValue")).toBe('{"test":true}');
  expect(onChange).toHaveBeenCalledWith({ test: true });
});

it("should not trigger the onChange handler if invalid JSON", () => {
  const schema: JSONSchema6 = {
    additionalProperties: {
      type: "string",
    },
    description: "Tags to be applied to new resources, specified as key/value pairs.",
    type: "object",
  };
  const onChange = jest.fn();
  const props = {
    onChange: (value: any) => {
      onChange(value);
    },
    schema,
  } as FieldProps;
  const wrapper = shallow(<CustomObjectField {...props} />);

  const editor = wrapper.find(AceEditor);
  editor.simulate("change", "test");
  expect(wrapper.state("rawValue")).toBe("test");
  expect(onChange).not.toHaveBeenCalled();
});
