// WARN: yaml doesn't have updated definitions for TypeScript
// In particular, it doesn't contain definitions for `get` and `set`
// that are used in this package
import * as AJV from "ajv";
import * as jsonSchema from "json-schema";
import { set } from "lodash";
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
): IBasicFormParam[] {
  let params: IBasicFormParam[] = [];
  if (schema && schema.properties) {
    const properties = schema.properties!;
    Object.keys(properties).map(propertyKey => {
      // The param path is its parent path + the object key
      const itemPath = `${parentPath || ""}${propertyKey}`;
      const { type, form } = properties[propertyKey];
      // If the property has the key "form", it's a basic parameter
      if (form) {
        // Use the default value either from the JSON schema or the default values
        const value = getValue(defaultValues, itemPath, properties[propertyKey].default);
        const param: IBasicFormParam = {
          ...properties[propertyKey],
          path: itemPath,
          type,
          value,
          children:
            properties[propertyKey].type === "object"
              ? retrieveBasicFormParams(defaultValues, properties[propertyKey], `${itemPath}.`)
              : undefined,
        };
        params = params.concat(param);
      } else {
        // If the property is an object, iterate recursively
        if (schema.properties![propertyKey].type === "object") {
          params = params.concat(
            retrieveBasicFormParams(defaultValues, properties[propertyKey], `${itemPath}.`),
          );
        }
      }
    });
  }
  return params;
}

function getDefinedPath(allElementsButTheLast: string[], doc: YAML.ast.Document) {
  let currentPath: string[] = [];
  let foundUndefined = false;
  allElementsButTheLast.forEach(p => {
    // Iterate over the path until finding an element that is not defined
    if (!foundUndefined) {
      const pathToEvaluate = currentPath.concat(p);
      const elem = (doc as any).getIn(pathToEvaluate);
      if (elem === undefined || elem === null) {
        foundUndefined = true;
      } else {
        currentPath = pathToEvaluate;
      }
    }
  });
  return currentPath;
}

// setValue modifies the current values (text) based on a path
export function setValue(values: string, path: string, newValue: any) {
  const doc = YAML.parseDocument(values);
  let splittedPath = path.split(".");
  // If the path is not defined (the parent nodes are undefined)
  // We need to change the path and the value to set to avoid accessing
  // the undefined node. For example, if a.b is undefined:
  // path: a.b.c, value: 1 ==> path: a.b, value: {c: 1}
  // TODO(andresmgot): In the future, this may be implemented in the YAML library itself
  // https://github.com/eemeli/yaml/issues/131
  const allElementsButTheLast = splittedPath.slice(0, splittedPath.length - 1);
  const parentNode = (doc as any).getIn(allElementsButTheLast);
  if (parentNode === undefined) {
    const definedPath = getDefinedPath(allElementsButTheLast, doc);
    const remainingPath = splittedPath.slice(definedPath.length + 1);
    newValue = set({}, remainingPath.join("."), newValue);
    splittedPath = splittedPath.slice(0, definedPath.length + 1);
  }
  (doc as any).setIn(splittedPath, newValue);
  return doc.toString();
}

// getValue returns the current value of an object based on YAML text and its path
export function getValue(values: string, path: string, defaultValue?: any) {
  const doc = YAML.parseDocument(values);
  const splittedPath = path.split(".");
  const value = (doc as any).getIn(splittedPath);
  return value === undefined || value === null ? defaultValue : value;
}

export function validate(
  values: string,
  schema: jsonSchema.JSONSchema4,
): { valid: boolean; errors: AJV.ErrorObject[] | null | undefined } {
  const ajv = new AJV();
  const valid = ajv.validate(schema, YAML.parse(values));
  return { valid: !!valid, errors: ajv.errors };
}
