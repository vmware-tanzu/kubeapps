// Copyright 2022-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import { CdsControlMessage } from "@cds/react/forms";
import { CdsIcon } from "@cds/react/icon";
import { CdsInput } from "@cds/react/input";
import { CdsRange } from "@cds/react/range";
import { CdsSelect } from "@cds/react/select";
import { CdsToggle } from "@cds/react/toggle";
import Column from "components/Column";
import Row from "components/Row";
import { isEmpty } from "lodash";
import { useState } from "react";
import { validateValuesSchema } from "shared/schema";
import { IAjvValidateResult, IBasicFormParam } from "shared/types";
import {
  basicFormsDebounceTime,
  getOptionalMin,
  getStringValue,
  getValueFromString,
} from "shared/utils";

export interface IArrayParamProps {
  id: string;
  label: string;
  type: string;
  param: IBasicFormParam;
  handleBasicFormParamChange: (
    param: IBasicFormParam,
  ) => (e: React.FormEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => void;
}

const getDefaultDataFromType = (type: string) => {
  switch (type) {
    case "number":
    case "integer":
      return 0;
    case "boolean":
      return false;
    case "object":
      return {};
    case "array":
      return [];
    case "string":
    default:
      return "";
  }
};

type supportedTypes = string | number | boolean | object | Array<any>;

export default function ArrayParam(props: IArrayParamProps) {
  const { id, label, type, param, handleBasicFormParamChange } = props;

  const initCurrentValue = () => {
    const currentValueInit = [];
    if (param.minItems) {
      for (let index = 0; index < param.minItems; index++) {
        currentValueInit[index] = getDefaultDataFromType(type);
      }
    }
    return currentValueInit;
  };

  const [currentArrayItems, setCurrentArrayItems] = useState<supportedTypes[]>(() => {
    return initCurrentValue();
  });
  const [validated, setValidated] = useState<IAjvValidateResult>();
  const [timeout, setThisTimeout] = useState({} as NodeJS.Timeout);

  const setArrayChangesInParam = () => {
    clearTimeout(timeout);
    const func = handleBasicFormParamChange(param);
    // The reference to target get lost, so we need to keep a copy
    const targetCopy = {
      currentTarget: {
        value: getStringValue(currentArrayItems),
        type: "array",
      },
    } as React.FormEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>;
    setThisTimeout(setTimeout(() => func(targetCopy), basicFormsDebounceTime));
  };

  const onChangeArrayItem = (
    e: React.FormEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>,
    index: number,
    value: supportedTypes,
  ) => {
    currentArrayItems[index] = value;
    setCurrentArrayItems([...currentArrayItems]);
    setArrayChangesInParam();

    // twofold validation: using the json schema (with ajv) and the html5 validation
    setValidated(validateValuesSchema(getStringValue(currentArrayItems), param.schema));
    e.currentTarget?.reportValidity();
  };

  const renderControlMsg = () =>
    !validated?.valid &&
    !isEmpty(validated?.errors) && (
      <>
        <CdsControlMessage status="error">
          {validated?.errors?.map((e: any) => e?.message).join(", ")}
        </CdsControlMessage>
        <br />
      </>
    );

  const step = param?.multipleOf || (type === "number" ? 0.5 : 1);

  const renderInput = (type: string, index: number) => {
    if (!isEmpty(param?.items?.enum)) {
      return (
        <>
          <CdsSelect layout="horizontal">
            <select
              required={param.isRequired}
              disabled={param.readOnly}
              aria-label={label}
              id={id}
              value={currentArrayItems[index] as string}
              onChange={e => onChangeArrayItem(e, index, e.currentTarget.value)}
            >
              <option disabled={true} key={""}>
                {""}
              </option>
              {param?.items?.enum?.map((enumValue: any) => (
                <option value={getValueFromString(enumValue)} key={enumValue}>
                  {enumValue}
                </option>
              ))}
            </select>
            {renderControlMsg()}
          </CdsSelect>
        </>
      );
    } else {
      switch (type) {
        case "number":
        case "integer":
          return (
            <>
              <CdsInput className="self-center">
                <input
                  required={param.isRequired}
                  disabled={param.readOnly}
                  min={getOptionalMin(param.exclusiveMinimum, param.minimum)}
                  max={getOptionalMin(param.exclusiveMaximum, param.maximum)}
                  aria-label={label}
                  id={`${id}-${index}_text`}
                  type="number"
                  onChange={e => onChangeArrayItem(e, index, Number(e.currentTarget.value))}
                  value={Number(currentArrayItems[index])}
                  step={step}
                />
              </CdsInput>
              <CdsRange>
                <input
                  required={param.isRequired}
                  disabled={param.readOnly}
                  min={getOptionalMin(param.exclusiveMinimum, param.minimum)}
                  max={getOptionalMin(param.exclusiveMaximum, param.maximum)}
                  aria-label={label}
                  id={`${id}-${index}_range`}
                  type="range"
                  onChange={e => onChangeArrayItem(e, index, Number(e.currentTarget.value))}
                  value={Number(currentArrayItems[index])}
                  step={step}
                />
              </CdsRange>
            </>
          );
        case "boolean":
          return (
            <CdsToggle>
              <input
                required={param.isRequired}
                disabled={param.readOnly}
                aria-label={label}
                id={`${id}-${index}_toggle`}
                type="checkbox"
                onChange={e => onChangeArrayItem(e, index, e.currentTarget.checked)}
                checked={!!currentArrayItems[index]}
              />
            </CdsToggle>
          );
        case "object":
          return (
            <CdsInput>
              <input
                required={param.isRequired}
                disabled={param.readOnly}
                aria-label={label}
                value={getStringValue(currentArrayItems[index])}
                onChange={e =>
                  onChangeArrayItem(e, index, getValueFromString(e.currentTarget.value, "object"))
                }
              />
            </CdsInput>
          );
        case "array":
          return (
            <CdsInput>
              <input
                required={param.isRequired}
                disabled={param.readOnly}
                aria-label={label}
                value={getStringValue(currentArrayItems[index])}
                onChange={e =>
                  onChangeArrayItem(e, index, getValueFromString(e.currentTarget.value, "array"))
                }
              />
            </CdsInput>
          );
        case "string":
        default:
          return (
            <CdsInput>
              <input
                required={param.isRequired}
                disabled={param.readOnly}
                maxLength={param.maxLength}
                minLength={param.minLength}
                pattern={param.pattern}
                aria-label={label}
                value={currentArrayItems[index] as string}
                onChange={e => onChangeArrayItem(e, index, e.currentTarget.value)}
              />
            </CdsInput>
          );
      }
    }
  };

  const onAddArrayItem = () => {
    currentArrayItems.push(getDefaultDataFromType(type));
    setCurrentArrayItems([...currentArrayItems]);
    setArrayChangesInParam();
  };

  const onDeleteArrayItem = (index: number) => {
    currentArrayItems.splice(index, 1);
    setCurrentArrayItems([...currentArrayItems]);
  };

  return (
    <>
      <CdsButton
        title={"Add a new value"}
        type="button"
        onClick={onAddArrayItem}
        action="flat"
        status="primary"
        size="sm"
        disabled={currentArrayItems.length >= param?.maxItems}
      >
        <CdsIcon shape="plus" size="sm" solid={true} />
        <span>Add</span>
      </CdsButton>
      {renderControlMsg()}
      {typeof currentArrayItems["map"] === "function" &&
        currentArrayItems?.map((_, index) => (
          <Row key={`${id}-${index}`}>
            <Column span={9}>{renderInput(type, index)}</Column>
            <Column span={1}>
              <CdsButton
                title={"Delete"}
                type="button"
                onClick={() => onDeleteArrayItem(index)}
                action="flat"
                status="primary"
                size="sm"
              >
                <CdsIcon shape="minus" size="sm" solid={true} />
              </CdsButton>
            </Column>
          </Row>
        ))}
    </>
  );
}
