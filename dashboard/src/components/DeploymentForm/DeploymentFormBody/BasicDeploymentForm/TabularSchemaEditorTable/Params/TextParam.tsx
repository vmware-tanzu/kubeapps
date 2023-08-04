// Copyright 2019-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsControlMessage } from "@cds/react/forms";
import { CdsInput } from "@cds/react/input";
import { CdsSelect } from "@cds/react/select";
import { CdsTextarea } from "@cds/react/textarea";
import Column from "components/Column";
import Row from "components/Row";
import { isEmpty } from "lodash";
import { useState } from "react";
import { validateValuesSchema } from "shared/schema";
import { IAjvValidateResult, IBasicFormParam } from "shared/types";
import { basicFormsDebounceTime, getStringValue, getValueFromString } from "shared/utils";

export interface ITextParamProps {
  id: string;
  label: string;
  inputType?: "text" | "textarea" | "password" | string;
  param: IBasicFormParam;
  handleBasicFormParamChange: (
    param: IBasicFormParam,
  ) => (e: React.FormEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => void;
}

export default function TextParam(props: ITextParamProps) {
  const { id, label, inputType, param, handleBasicFormParamChange } = props;

  const [validated, setValidated] = useState<IAjvValidateResult>();
  const [currentValue, setCurrentValue] = useState(getStringValue(param.currentValue));
  const [isValueModified, setIsValueModified] = useState(false);
  const [timeout, setThisTimeout] = useState({} as NodeJS.Timeout);

  const onChange = (
    e: React.FormEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>,
  ) => {
    // update the current value
    setCurrentValue(e.currentTarget.value);
    setIsValueModified(
      getStringValue(e.currentTarget.value) !== getStringValue(param.currentValue),
    );

    // twofold validation: using the json schema (with ajv) and the html5 validation
    setValidated(validateValuesSchema(e.currentTarget.value, param.schema));
    e.currentTarget?.reportValidity();

    // Gather changes before submitting
    clearTimeout(timeout);
    const func = handleBasicFormParamChange(param);
    // The reference to target get lost, so we need to keep a copy
    const targetCopy = {
      currentTarget: {
        value: getStringValue(e.currentTarget?.value),
        type: param.type === "object" ? param.type : e.currentTarget?.type,
      },
    } as React.FormEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>;
    setThisTimeout(setTimeout(() => func(targetCopy), basicFormsDebounceTime));
  };

  const unsavedMessage = isValueModified ? "Unsaved" : "";
  const isDiffCurrentVsDefault =
    getStringValue(param.currentValue, param.type) !==
    getStringValue(param.defaultValue, param.type);
  const isDiffCurrentVsDeployed =
    getStringValue(param.currentValue, param.type) !==
    getStringValue(param.defaultValue, param.type);
  const isModified =
    isValueModified ||
    (isDiffCurrentVsDefault && (!param.deployedValue || isDiffCurrentVsDeployed));

  const renderControlMsg = () =>
    !validated?.valid && !isEmpty(validated?.errors) ? (
      <>
        <CdsControlMessage status="error">
          {unsavedMessage}
          <br />
          {validated?.errors?.map(e => e.message).join(", ")}
        </CdsControlMessage>
        <br />
      </>
    ) : (
      <CdsControlMessage>{unsavedMessage}</CdsControlMessage>
    );

  const renderInput = () => {
    if (!isEmpty(param.enum)) {
      return (
        <>
          <CdsSelect layout="horizontal">
            <select
              required={param.isRequired}
              disabled={param.readOnly}
              aria-label={label}
              id={id}
              value={getStringValue(currentValue, param.type)}
              onChange={onChange}
            >
              <option disabled={true} key={""}>
                {""}
              </option>
              {param?.enum?.map((enumValue: any) => (
                <option value={getValueFromString(enumValue)} key={enumValue}>
                  {enumValue}
                </option>
              ))}
            </select>
            {renderControlMsg()}
          </CdsSelect>
        </>
      );
    } else if (param.type === "string") {
      switch (inputType) {
        case undefined:
        case "textarea":
          return (
            <CdsTextarea className={isModified ? "bolder" : ""}>
              <textarea
                required={param.isRequired}
                disabled={param.readOnly}
                maxLength={param.maxLength}
                minLength={param.minLength}
                aria-label={label}
                id={id}
                value={getStringValue(currentValue, param.type)}
                onChange={onChange}
              />
              {renderControlMsg()}
            </CdsTextarea>
          );
        default:
          return (
            <>
              <datalist id={`${id}-examples`}>
                {param?.examples?.map((example: any) => (
                  <option value={getValueFromString(example)} key={example}>
                    {example}
                  </option>
                ))}
              </datalist>
              <CdsInput className={isModified ? "bolder" : ""}>
                <input
                  required={param.isRequired}
                  disabled={param.readOnly}
                  maxLength={param.maxLength}
                  minLength={param.minLength}
                  pattern={param.pattern}
                  aria-label={label}
                  id={id}
                  list={param.examples ? `${id}-examples` : ""}
                  type={inputType}
                  value={getStringValue(currentValue, param.type)}
                  onChange={onChange}
                />
                {renderControlMsg()}
              </CdsInput>
            </>
          );
      }
      // is an object
    } else {
      return (
        <>
          <CdsTextarea className={isModified ? "bolder" : ""}>
            <textarea
              required={param.isRequired}
              disabled={param.readOnly}
              maxLength={param.maxLength}
              minLength={param.minLength}
              aria-label={label}
              id={id}
              value={getStringValue(currentValue)}
              onChange={onChange}
            />
            {renderControlMsg()}
          </CdsTextarea>
        </>
      );
    }
  };

  return (
    <Row>
      <Column span={10}>{renderInput()}</Column>
    </Row>
  );
}
