// WARN: yaml doesn't have updated definitions for TypeScript
import * as YAML from "yaml";
import { IBasicFormParam } from "./types";

// Avoid to explicitly add "null" when an element is not defined
import { nullOptions } from "yaml/types";
nullOptions.nullStr = "";

// retrieveBasicFormParams iterates over a JSON Schema properties looking for `form` keys
// It uses the raw yaml to setup default values.
// It returns a key:value map for easier handling.
export function retrieveBasicFormParams(
  defaultValues: string,
  schema: any,
): { [key: string]: IBasicFormParam } {
  if (schema && schema.properties) {
    return lookForFormParams(defaultValues, schema.properties);
  }
  return {};
}

// getDocumentElem returns the YAML element given a root element
export function getDocumentElem(rootElem: any, path: string): any | undefined {
  // Split the path per keys
  const splittedPath = path.split(".");
  if (splittedPath.length === 1) {
    // If the path doesn't contain inner objects, simply return
    return rootElem.get(path);
  }
  // If the path contains inner elements, iterate until finding the last one
  let res = rootElem.get(splittedPath[0]);
  splittedPath.slice(1, splittedPath.length).forEach(p => (res = res && res.get(p)));
  return res;
}

// setValue modifies the current values (text) based on a path
export function setValue(values: string, path: string, newValue: any) {
  const doc = YAML.parseDocument(values);
  if (path.includes(".")) {
    // If the path points to an inner element we need to parse it
    // and obtain the previous element to set it
    // e.g.
    // path = "a.b.c"
    // itemToSet = "c"
    // elem = "a.b"
    // elem.set(itemToSet, value)
    const splittedPath = path.split(".");
    const itemToSet = splittedPath[splittedPath.length - 1];
    splittedPath.pop();
    const pathToSet = splittedPath.join(".");
    const elem = getDocumentElem(doc, pathToSet);
    if (!elem) {
      // TODO: It should be possible to handle this situation and add a new object
      throw new Error(`Unable to set the value ${path}. Parent element not found`);
    }
    elem.set(itemToSet, newValue);
  } else {
    doc.set(path, newValue);
  }
  return doc.toString();
}

// getDefaultValues returns the current value of an object based on YAML text and its path
export function getDefaultValue(defaultValues: string, path: string): string | undefined {
  const doc = YAML.parseDocument(defaultValues);
  if (path.includes(".")) {
    const elem = getDocumentElem(doc, path);
    // If the path is found, return it as a string
    return elem && elem.toString();
  }
  return path.includes(".") ? getDocumentElem(doc, path).toString() : doc.get(path);
}

// lookForFormParams iterates recursively over JSON Schema properties and returns IBasicFormParams
function lookForFormParams(
  defaultValues: string,
  schemaProperties: {},
  parentPath?: string,
  params?: { [name: string]: IBasicFormParam },
) {
  if (!params) {
    params = {};
  }
  Object.keys(schemaProperties).map(propertyKey => {
    // The param path is its parent path + the object key
    const itemPath = `${parentPath || ""}${propertyKey}`;
    // If the property has the key "form", it's a basic parameter
    if (schemaProperties[propertyKey].form) {
      // The key of the param is the value of the form tag
      params![schemaProperties[propertyKey].form] = {
        path: itemPath,
        // Use the default value either from the JSON schema or the default values
        value: getDefaultValue(defaultValues, itemPath) || schemaProperties[propertyKey].default,
      };
    }
    // If the property is an object, iterate recursively
    if (schemaProperties[propertyKey].type === "object") {
      params = lookForFormParams(
        defaultValues,
        schemaProperties[propertyKey].properties,
        `${itemPath}.`,
        params,
      );
    }
  });
  return params;
}
