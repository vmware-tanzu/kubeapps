// Copyright 2019-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsControlMessage } from "@cds/react/forms";
import { CdsInput } from "@cds/react/input";
import { CdsRange } from "@cds/react/range";
import Column from "components/Column";
import Row from "components/Row";
import { isEmpty, toNumber } from "lodash";
import { useState } from "react";
import { validateValuesSchema } from "shared/schema";
import { IAjvValidateResult, IBasicFormParam } from "shared/types";
import { basicFormsDebounceTime, getOptionalMin, getStringValue } from "shared/utils";

export interface ISliderParamProps {
  id: string;
  label: string;
  param: IBasicFormParam;
  unit: string;
  handleBasicFormParamChange: (
    p: IBasicFormParam,
  ) => (e: React.FormEvent<HTMLInputElement>) => void;
}

export default function SliderParam(props: ISliderParamProps) {
  const { handleBasicFormParamChange, id, label, param } = props;

  const initCurrentValue = () =>
    toNumber(param.currentValue) ||
    toNumber(param.exclusiveMinimum) ||
    toNumber(param.minimum) ||
    0;

  const [validated, setValidated] = useState<IAjvValidateResult>();
  const [currentValue, setCurrentValue] = useState(initCurrentValue());

  const [isValueModified, setIsValueModified] = useState(false);
  const [timeout, setThisTimeout] = useState({} as NodeJS.Timeout);

  const step =
    param?.multipleOf || (param.schema?.type === "number" || param?.type === "number" ? 0.5 : 1);

  const onChange = (e: React.FormEvent<HTMLInputElement>) => {
    setCurrentValue(toNumber(e.currentTarget.value));
    setIsValueModified(toNumber(e.currentTarget.value) !== toNumber(param.currentValue));

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
        type: e.currentTarget?.type,
      },
    } as React.FormEvent<HTMLInputElement>;
    setThisTimeout(setTimeout(() => func(targetCopy), basicFormsDebounceTime));
  };

  const unsavedMessage = isValueModified ? "Unsaved" : "";
  const isModified =
    isValueModified ||
    (param.currentValue !== param.defaultValue && param.currentValue !== param.deployedValue);

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

  const input = (
    <CdsRange>
      <input
        required={param.isRequired}
        disabled={param.readOnly}
        min={getOptionalMin(param.exclusiveMinimum, param.minimum)}
        max={getOptionalMin(param.exclusiveMaximum, param.maximum)}
        aria-label={label}
        id={id + "_range"}
        type="range"
        step={step}
        onChange={onChange}
        value={currentValue}
      />
      {renderControlMsg()}
    </CdsRange>
  );

  const inputText = (
    <div className="self-center">
      <CdsInput className={isModified ? "bolder" : ""}>
        <input
          required={param.isRequired}
          disabled={param.readOnly}
          min={getOptionalMin(param.exclusiveMinimum, param.minimum)}
          max={getOptionalMin(param.exclusiveMaximum, param.maximum)}
          aria-label={label}
          id={id + "_text"}
          type="number"
          step={step}
          onChange={onChange}
          value={currentValue}
        />
      </CdsInput>
    </div>
  );

  return (
    <Row>
      <Column span={4}>{inputText}</Column>
      <Column span={6}>{input}</Column>
    </Row>
  );
}
