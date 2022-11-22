// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import Ajv, { ErrorObject, JSONSchemaType } from "ajv";
import { findIndex, isEmpty, set } from "lodash";
import { DeploymentEvent, IAjvValidateResult, IBasicFormParam } from "shared/types";
import YAML from "yaml";
import * as jsonpatch from "fast-json-patch";
import { getPathValueInYamlNode, getPathValueInYamlNodeWithDefault } from "./yamlUtils";

const ajv = new Ajv({ strict: false });

const IS_CUSTOM_COMPONENT_PROP_NAME = "x-is-custom-component";

export function retrieveBasicFormParams(
  currentValues: YAML.Document.Parsed<YAML.ParsedNode>,
  packageValues: YAML.Document.Parsed<YAML.ParsedNode>,
  schema: JSONSchemaType<any>,
  deploymentEvent: DeploymentEvent,
  deployedValues?: YAML.Document.Parsed<YAML.ParsedNode>,
  parentPath?: string,
): IBasicFormParam[] {
  let params: IBasicFormParam[] = [];
  if (schema?.properties && !isEmpty(schema.properties)) {
    const properties = schema.properties;
    const schemaExamples = schema.examples;
    Object.keys(properties).forEach(propertyKey => {
      const schemaProperty = properties[propertyKey] as JSONSchemaType<any>;
      // The param path is its parent path + the object key
      const itemPath = `${parentPath || ""}${propertyKey}`;
      const isUpgrading = deploymentEvent === "upgrade" && deployedValues;
      const isLeaf = !schemaProperty?.properties;

      // get the values for the current property in the examples array
      // for objects, we need to get the value of the property in the example array,
      // for the rest, we can just get the value of the example array
      let examples = schemaProperty.examples;
      if (schemaExamples?.length > 0) {
        examples = schemaExamples?.map((item: any) =>
          typeof item === "object" ? item?.[propertyKey]?.toString() ?? "" : item?.toString() ?? "",
        );
      }

      const param: IBasicFormParam = {
        ...schemaProperty,
        title: schemaProperty.title || propertyKey,
        key: itemPath,
        schema: schemaProperty,
        hasProperties: Boolean(schemaProperty?.properties),
        params: schemaProperty?.properties
          ? retrieveBasicFormParams(
              currentValues,
              packageValues,
              schemaProperty,
              deploymentEvent,
              deployedValues,
              `${itemPath}/`,
            )
          : undefined,
        // get the string values of the enum array
        enum: schemaProperty?.enum?.map((item: { toString: () => any }) => item?.toString() ?? ""),
        // We leave the validation of user values to the Helm backend, since it will take
        // into account default values (on install) and previously set values (on upgrade).
        isRequired: false,
        examples: examples,
        // If exists, the value that is currently deployed
        deployedValue: isLeaf
          ? isUpgrading
            ? getPathValueInYamlNode(deployedValues, itemPath)
            : ""
          : "",
        // The default is the value coming from the package values or the one defined in the schema,
        // or vice-verse, which one should take precedence?
        defaultValue: isLeaf
          ? getPathValueInYamlNodeWithDefault(packageValues, itemPath, schemaProperty.default)
          : "",
        // same as default value, but this one will be later overwritten by the user input
        currentValue: isLeaf
          ? getPathValueInYamlNodeWithDefault(currentValues, itemPath, schemaProperty.default)
          : "",
        isCustomComponent:
          schemaProperty?.customComponent || schemaProperty?.[IS_CUSTOM_COMPONENT_PROP_NAME],
      };
      params = params.concat(param);

      if (!schemaProperty?.properties) {
        params = params.concat(
          retrieveBasicFormParams(
            currentValues,
            packageValues,
            schemaProperty,
            deploymentEvent,
            deployedValues,
            `${itemPath}/`,
          ),
        );
      }
    });
  }
  return params;
}

export function updateCurrentConfigByKey(
  paramsList: IBasicFormParam[],
  key: string,
  value: any,
  depth = 1,
): any {
  if (!paramsList) {
    return [];
  }

  // Find item index using findIndex
  const indexLeaf = findIndex(paramsList, { key: key });
  // is it a leaf node?
  if (!paramsList?.[indexLeaf]) {
    const a = key.split("/").slice(0, depth).join("/");
    const index = findIndex(paramsList, { key: a });
    if (paramsList?.[index]?.params) {
      set(
        paramsList[index],
        "currentValue",
        updateCurrentConfigByKey(paramsList?.[index]?.params || [], key, value, depth + 1),
      );
      return paramsList;
    }
  }
  // Replace item at index using native splice
  paramsList?.splice(indexLeaf, 1, {
    ...paramsList[indexLeaf],
    currentValue: value,
  });
  return paramsList;
}

export function schemaToString(schema: JSONSchemaType<any> | undefined): string {
  let schemaString;
  try {
    schemaString = JSON.stringify(schema, null, 2);
  } catch (e) {
    schemaString = "{}";
  }
  return schemaString;
}

export function schemaToObject(schema?: string): JSONSchemaType<any> {
  let schemaObject;
  try {
    schemaObject = JSON.parse(schema || "{}");
  } catch (e) {
    schemaObject = {};
  }
  return schemaObject as JSONSchemaType<any>;
}

export function validateValuesSchema(
  values: string,
  schema: JSONSchemaType<any> | any,
  defaultValues?: string,
): { valid: boolean; errors: ErrorObject[] | null | undefined } {
  let valuesToCheck = YAML.parse(values);
  if (defaultValues) {
    const defaultYAML = YAML.parse(defaultValues);
    let patches = jsonpatch.compare(defaultYAML as any, valuesToCheck as any);
    patches = patches.filter(function (d) {
      return ["add", "replace"].includes(d.op);
    });
    valuesToCheck = jsonpatch.applyPatch(defaultYAML, patches).newDocument;
  }
  const valid = ajv.validate(schema, valuesToCheck);
  return { valid: !!valid, errors: ajv.errors } as IAjvValidateResult;
}

export function validateSchema(schema: JSONSchemaType<any>): IAjvValidateResult {
  const valid = ajv.validateSchema(schema);
  return { valid: valid, errors: ajv.errors } as IAjvValidateResult;
}
