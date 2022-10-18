// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsControlMessage } from "@cds/react/forms";
import { CdsInput } from "@cds/react/input";
import { CdsSelect } from "@cds/react/select";
import { CdsTextarea } from "@cds/react/textarea";
import Column from "components/js/Column";
import Row from "components/js/Row";
import { isEmpty } from "lodash";
import { useState } from "react";
import { validateValuesSchema } from "shared/schema";
import { IAjvValidateResult, IBasicFormParam } from "shared/types";
import { basicFormsDebounceTime } from "shared/utils";

export interface ITextParamProps {
  id: string;
  label: string;
  inputType?: "text" | "textarea" | string;
  param: IBasicFormParam;
  handleBasicFormParamChange: (
    param: IBasicFormParam,
  ) => (e: React.FormEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => void;
}

function getStringValue(param: IBasicFormParam, value?: any) {
  if (["array", "object"].includes(param?.type)) {
    return JSON.stringify(value || param?.currentValue);
  } else {
    return value?.toString() || param?.currentValue?.toString();
  }
}
function getValueFromString(param: IBasicFormParam, value: any) {
  if (["array", "object"].includes(param?.type)) {
    try {
      return JSON.parse(value);
    } catch (e) {
      return value?.toString();
    }
  } else {
    return value?.toString();
  }
}

function toStringValue(value: any) {
  return JSON.stringify(value?.toString() || "");
}

export default function TextParam(props: ITextParamProps) {
  const { id, label, inputType, param, handleBasicFormParamChange } = props;

  const [validated, setValidated] = useState<IAjvValidateResult>();
  const [currentValue, setCurrentValue] = useState(getStringValue(param));
  const [isValueModified, setIsValueModified] = useState(false);
  const [timeout, setThisTimeout] = useState({} as NodeJS.Timeout);

  const onChange = (
    e: React.FormEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>,
  ) => {
    setValidated(validateValuesSchema(e.currentTarget.value, param.schema));
    setCurrentValue(e.currentTarget.value);
    setIsValueModified(toStringValue(e.currentTarget.value) !== toStringValue(param.currentValue));
    // Gather changes before submitting
    clearTimeout(timeout);
    const func = handleBasicFormParamChange(param);
    // The reference to target get lost, so we need to keep a copy
    const targetCopy = {
      currentTarget: {
        value: getValueFromString(param, e.currentTarget?.value),
        type: e.currentTarget?.type,
      },
    } as React.FormEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>;
    setThisTimeout(setTimeout(() => func(targetCopy), basicFormsDebounceTime));
  };

  const unsavedMessage = isValueModified ? "Unsaved" : "";
  const isDiffCurrentVsDefault =
    toStringValue(param.currentValue) !== toStringValue(param.defaultValue);
  const isDiffCurrentVsDeployed =
    toStringValue(param.currentValue) !== toStringValue(param.defaultValue);
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

  let input = (
    <>
      <CdsInput className={isModified ? "bolder" : ""}>
        <input
          required={param.required}
          aria-label={label}
          id={id}
          type={inputType ?? "text"}
          value={currentValue ?? ""}
          onChange={onChange}
        />
        {renderControlMsg()}
      </CdsInput>
    </>
  );
  if (inputType === "textarea") {
    input = (
      <CdsTextarea className={isModified ? "bolder" : ""}>
        <textarea
          required={param.required}
          aria-label={label}
          id={id}
          value={currentValue ?? ""}
          onChange={onChange}
        />
        {renderControlMsg()}
      </CdsTextarea>
    );
  } else if (!isEmpty(param.enum)) {
    input = (
      <>
        <CdsSelect layout="horizontal" className={isModified ? "bolder" : ""}>
          <select
            required={param.required}
            aria-label={label}
            id={id}
            onChange={onChange}
            value={currentValue}
          >
            {param?.enum?.map((enumValue: any) => (
              <option key={enumValue}>{enumValue}</option>
            ))}
          </select>
          {renderControlMsg()}
        </CdsSelect>
      </>
    );
  }

  return (
    <Row>
      <Column span={10}>{input}</Column>
    </Row>
  );
}
