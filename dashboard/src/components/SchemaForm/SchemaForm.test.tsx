import { shallow } from "enzyme";
import { JSONSchema6 } from "json-schema";
import * as React from "react";
import Form from "react-jsonschema-form";

import ArrayFieldTemplate from "./ArrayFieldTemplate";
import CustomObjectField from "./CustomObjectField";
import FieldTemplate from "./FieldTemplate";
import SchemaForm from "./SchemaForm";

it("renders the form with the schema and templates", () => {
  const schema: JSONSchema6 = {
    properties: {
      age: {
        description: "Age in years",
        minimum: 0,
        type: "integer",
      },
      firstName: {
        type: "string",
      },
      lastName: {
        type: "string",
      },
    },
    required: ["firstName", "lastName"],
    title: "Person",
    type: "object",
  };
  const wrapper = shallow(<SchemaForm schema={schema} />);
  const form = wrapper.find(Form);
  expect(form.exists()).toBe(true);
  expect(form.props().schema).toEqual(schema);
  expect(form.props().autocomplete).toBe("off");
  expect(form.props().FieldTemplate).toBe(FieldTemplate);
  expect(form.props().ArrayFieldTemplate).toBe(ArrayFieldTemplate);
  expect(form.props().fields).toEqual({ ObjectField: CustomObjectField });
});

it("renders children", () => {
  const schema: JSONSchema6 = {
    properties: {
      firstName: {
        type: "string",
      },
    },
    type: "object",
  };
  const button = (
    <button className="button button-primary" type="submit">
      Submit
    </button>
  );
  const wrapper = shallow(<SchemaForm schema={schema}>{button}</SchemaForm>);
  expect(wrapper.find("button").exists()).toBe(true);
  expect(wrapper.find("button").text()).toBe("Submit");
});

it("takes an onSubmit handler", () => {
  const schema: JSONSchema6 = {
    properties: {
      firstName: {
        type: "string",
      },
    },
    type: "object",
  };
  const onSubmit = jest.fn();
  const wrapper = shallow(<SchemaForm schema={schema} onSubmit={onSubmit} />);
  wrapper.find(Form).simulate("submit");
  expect(onSubmit).toHaveBeenCalled();
});
