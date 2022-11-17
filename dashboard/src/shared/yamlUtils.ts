// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { unescapePathComponent } from "fast-json-patch";
import { isEmpty, set } from "lodash";
import YAML, { Scalar, ToStringOptions } from "yaml";
import { getStringValue } from "./utils";

const toStringOptions: ToStringOptions = {
  defaultKeyType: "PLAIN",
  defaultStringType: Scalar.QUOTE_DOUBLE, // Preserving double quotes in scalars (see https://github.com/vmware-tanzu/kubeapps/issues/3621)
  nullStr: "", // Avoid to explicitly add "null" when an element is not defined
};

export function parseToYamlNode(string: string) {
  return YAML.parseDocument(string, { toStringDefaults: toStringOptions });
}

export function parseObjectToYamlNode(object: object) {
  return YAML.parseDocument(getStringValue(object), { toStringDefaults: toStringOptions });
}

export function parseToJS(string?: string) {
  return parseToYamlNode(string || "{}").toJS();
}

export function parseToString(object?: object | string) {
  return YAML.stringify(object, { toStringDefaults: toStringOptions });
}

export function toStringYamlNode(valuesNode: YAML.Document.Parsed<YAML.ParsedNode>) {
  return valuesNode.toString(toStringOptions);
}

export function setPathValueInYamlNode(
  valuesNode: YAML.Document.Parsed<YAML.ParsedNode>,
  path: string,
  newValue: any,
) {
  const { splitPath: split, value } = parsePathAndValue(valuesNode, path, newValue);
  valuesNode.setIn(split, value);
  return valuesNode;
}

function parsePathAndValue(doc: YAML.Document, path: string, value?: any) {
  if (isEmpty(doc.contents)) {
    // If the doc is empty we have an special case
    return { value: set({}, path.replace(/^\//, ""), value), splitPath: [] };
  }
  let splitPath = splitPathBySlash(path);
  // If the path is not defined (the parent nodes are undefined)
  // We need to change the path and the value to set to avoid accessing
  // the undefined node. For example, if a.b is undefined:
  // path: a.b.c, value: 1 ==> path: a.b, value: {c: 1}
  // TODO(andresmgot): In the future, this may be implemented in the YAML library itself
  // https://github.com/eemeli/yaml/issues/131
  const allElementsButTheLast = splitPath.slice(0, splitPath.length - 1);
  const parentNode = (doc as any).getIn(allElementsButTheLast);
  if (parentNode === undefined) {
    const definedPath = getDefinedPath(allElementsButTheLast, doc);
    const remainingPath = splitPath.slice(definedPath.length + 1);
    value = set({}, remainingPath.join("."), value);
    splitPath = splitPath.slice(0, definedPath.length + 1);
  }
  return { splitPath: unescapePath(splitPath), value };
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

export function getPathValueInYamlNodeWithDefault(
  values: YAML.Document.Parsed<YAML.ParsedNode>,
  path: string,
  defaultValue?: any,
) {
  const value = getPathValueInYamlNode(values, path);

  return value === undefined || value === null ? defaultValue : value;
}

export function getPathValueInYamlNode(
  values: YAML.Document.Parsed<YAML.ParsedNode>,
  path: string,
) {
  const splitPath = parsePath(path);
  let value = values?.getIn(splitPath);

  // if the value from getIn is an object, it means
  // it is a YamlSeq or YamlMap, so we need to convert it
  // back to a plain JS object
  if (typeof value === "object") {
    value = (value as any)?.toJSON();
  }

  return value;
}

function parsePath(path: string): string[] {
  return unescapePath(splitPathBySlash(path));
}

function unescapePath(path: string[]): string[] {
  // jsonpath escapes slashes to not mistake then with objects so we need to revert that
  return path.map(p => unescapePathComponent(p));
}

function splitPathBySlash(path: string): string[] {
  return (
    (path ?? "")
      // ignore the first slash, if exists
      .replace(/^\//, "")
      // split by slashes
      .split("/")
  );
}

export function deleteValue(values: string, path: string) {
  const doc = YAML.parseDocument(values, { toStringDefaults: toStringOptions });
  const { splitPath } = parsePathAndValue(doc, path);
  (doc as any).deleteIn(splitPath);
  // If the document is empty after the deletion instead of returning {}
  // we return an empty line "\n"
  return doc.contents && !isEmpty((doc.contents as any).items)
    ? doc.toString(toStringOptions)
    : "\n";
}

// setValue modifies the current values (text) based on a path
export function setValue(values: string, path: string, newValue: any) {
  const doc = YAML.parseDocument(values, { toStringDefaults: toStringOptions });
  const { splitPath, value } = parsePathAndValue(doc, path, newValue);
  (doc as any).setIn(splitPath, value);
  return doc.toString(toStringOptions);
}
