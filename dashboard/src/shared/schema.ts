// WARN: yaml doesn't have updated definitions for TypeScript
// In particular, it doesn't contain definitions for `get` and `set`
// that are used in this package
import * as jsonSchema from "json-schema";
import * as YAML from "yaml";
import { IBasicFormParam } from "./types";

// Avoid to explicitly add "null" when an element is not defined
// tslint:disable-next-line
const { nullOptions } = require("yaml/types");
nullOptions.nullStr = "";

// retrieveBasicFormParams iterates over a JSON Schema properties looking for `form` keys
// It uses the raw yaml to setup default values.
// It returns a key:value map for easier handling.
export function retrieveBasicFormParams(
  defaultValues: string,
  schema?: jsonSchema.JSONSchema4,
  parentPath?: string,
): { [key: string]: IBasicFormParam } {
  let params = {};
  if (schema && schema.properties) {
    const properties = schema.properties!;
    Object.keys(properties).map(propertyKey => {
      // The param path is its parent path + the object key
      const itemPath = `${parentPath || ""}${propertyKey}`;
      // If the property has the key "form", it's a basic parameter
      if (properties[propertyKey].form) {
        // Use the default value either from the JSON schema or the default values
        const value = getValue(defaultValues, itemPath, properties[propertyKey].default);
        params = {
          ...params,
          // The key of the param is the value of the form tag
          [properties[propertyKey].form]: { path: itemPath, value },
        };
      }
      // If the property is an object, iterate recursively
      if (schema.properties![propertyKey].type === "object") {
        params = {
          ...params,
          ...retrieveBasicFormParams(
            defaultValues,
            schema.properties![propertyKey],
            `${itemPath}.`,
          ),
        };
      }
    });
  }
  return params;
}

// setValue modifies the current values (text) based on a path
export function setValue(values: string, path: string, newValue: any) {
  const doc = YAML.parseDocument(values);
  const splittedPath = path.split(".");
  (doc as any).setIn(splittedPath, newValue);
  return doc.toString();
}

// getValue returns the current value of an object based on YAML text and its path
export function getValue(values: string, path: string, defaultValue?: any) {
  const doc = YAML.parseDocument(values);
  const splittedPath = path.split(".");
  return (doc as any).getIn(splittedPath) || defaultValue;
}
