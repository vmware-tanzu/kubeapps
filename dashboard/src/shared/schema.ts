// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import Ajv, { ErrorObject, JSONSchemaType } from "ajv";
import * as jsonpatch from "fast-json-patch";
import * as yaml from "js-yaml";
import { isEmpty, set } from "lodash";
// TODO(agamez): check if we can replace this package by js-yaml or vice-versa
import YAML, { ToStringOptions, Scalar } from "yaml";
import { IBasicFormParam } from "./types";

const ajv = new Ajv({ strict: false });

const toStringOptions: ToStringOptions = {
  defaultKeyType: "PLAIN",
  defaultStringType: Scalar.QUOTE_DOUBLE, // Preserving double quotes in scalars (see https://github.com/vmware-tanzu/kubeapps/issues/3621)
  nullStr: "", // Avoid to explicitly add "null" when an element is not defined
};

// retrieveBasicFormParams iterates over a JSON Schema properties looking for `form` keys
// It uses the raw yaml to setup default values.
// It returns a key:value map for easier handling.
export function retrieveBasicFormParams(
  defaultValues: string,
  schema?: JSONSchemaType<any>,
  parentPath?: string,
): IBasicFormParam[] {
  let params: IBasicFormParam[] = [];

  if (schema && schema.properties) {
    const properties = schema.properties!;
    Object.keys(properties).forEach(propertyKey => {
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
          enum: properties[propertyKey].enum?.map(
            (item: { toString: () => any }) => item?.toString() ?? "",
          ),
          children:
            properties[propertyKey].type === "object"
              ? retrieveBasicFormParams(defaultValues, properties[propertyKey], `${itemPath}/`)
              : undefined,
        };
        params = params.concat(param);
      } else {
        // If the property is an object, iterate recursively
        if (schema.properties![propertyKey].type === "object") {
          params = params.concat(
            retrieveBasicFormParams(defaultValues, properties[propertyKey], `${itemPath}/`),
          );
        }
      }
    });
  }
  return params;
}

function getDefinedPath(allElementsButTheLast: string[], doc: YAML.Document) {
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

function splitPath(path: string): string[] {
  return (
    (path ?? "")
      // ignore the first slash, if exists
      .replace(/^\//, "")
      // split by slashes
      .split("/")
  );
}

function unescapePath(path: string[]): string[] {
  // jsonpath escapes slashes to not mistake then with objects so we need to revert that
  return path.map(p => jsonpatch.unescapePathComponent(p));
}

function parsePath(path: string): string[] {
  return unescapePath(splitPath(path));
}

function parsePathAndValue(doc: YAML.Document, path: string, value?: any) {
  if (isEmpty(doc.contents)) {
    // If the doc is empty we have an special case
    return { value: set({}, path.replace(/^\//, ""), value), splittedPath: [] };
  }
  let splittedPath = splitPath(path);
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
    value = set({}, remainingPath.join("."), value);
    splittedPath = splittedPath.slice(0, definedPath.length + 1);
  }
  return { splittedPath: unescapePath(splittedPath), value };
}

// setValue modifies the current values (text) based on a path
export function setValue(values: string, path: string, newValue: any) {
  const doc = YAML.parseDocument(values, { toStringDefaults: toStringOptions });
  const { splittedPath, value } = parsePathAndValue(doc, path, newValue);
  (doc as any).setIn(splittedPath, value);
  return doc.toString(toStringOptions);
}

// parseValues returns a processed version of the values without modifying anything
export function parseValues(values: string) {
  return YAML.parseDocument(values, {
    toStringDefaults: toStringOptions,
  }).toString(toStringOptions);
}

export function deleteValue(values: string, path: string) {
  const doc = YAML.parseDocument(values, { toStringDefaults: toStringOptions });
  const { splittedPath } = parsePathAndValue(doc, path);
  (doc as any).deleteIn(splittedPath);
  // If the document is empty after the deletion instead of returning {}
  // we return an empty line "\n"
  return doc.contents && !isEmpty((doc.contents as any).items)
    ? doc.toString(toStringOptions)
    : "\n";
}

// getValue returns the current value of an object based on YAML text and its path
export function getValue(values: string, path: string, defaultValue?: any) {
  const doc = YAML.parseDocument(values, { toStringDefaults: toStringOptions });
  const splittedPath = parsePath(path);
  const value = (doc as any).getIn(splittedPath);
  return value === undefined || value === null ? defaultValue : value;
}

export function validate(
  values: string,
  schema: JSONSchemaType<any> | any,
): { valid: boolean; errors: ErrorObject[] | null | undefined } {
  const valid = ajv.validate(schema, yaml.load(values));
  return { valid: !!valid, errors: ajv.errors };
}
