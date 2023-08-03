// Copyright 2019-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsControlMessage } from "@cds/react/forms";
import { CdsToggle, CdsToggleGroup } from "@cds/react/toggle";
import Column from "components/Column";
import Row from "components/Row";
import { useState } from "react";
import { IBasicFormParam } from "shared/types";
import { getStringValue } from "shared/utils";

export interface IBooleanParamProps {
  id: string;
  label: string;
  param: IBasicFormParam;
  handleBasicFormParamChange: (
    p: IBasicFormParam,
  ) => (e: React.FormEvent<HTMLInputElement>) => void;
}
export default function BooleanParam(props: IBooleanParamProps) {
  const { id, label, param, handleBasicFormParamChange } = props;

  const [currentValue, setCurrentValue] = useState(param.currentValue || false);
  const [isValueModified, setIsValueModified] = useState(false);

  const onChange = (e: React.FormEvent<HTMLInputElement>) => {
    // create an event that "getValueFromEvent" can process,
    const event = {
      currentTarget: {
        //convert the boolean "checked" prop to a normal "value" string one
        value: getStringValue(e.currentTarget?.checked),
        type: "checkbox",
      },
    } as React.FormEvent<HTMLInputElement>;
    setCurrentValue(e.currentTarget?.checked);
    setIsValueModified(e.currentTarget?.checked !== param.currentValue);
    handleBasicFormParamChange(param)(event);
  };

  const unsavedMessage = isValueModified ? "Unsaved" : "";

  const isModified =
    isValueModified ||
    (param.currentValue !== param.defaultValue && param.currentValue !== param.deployedValue);

  const input = (
    <CdsToggleGroup id={id + "_group"}>
      <label htmlFor={id + "_group"}>{""}</label>
      <CdsToggle>
        <input
          required={param.isRequired}
          disabled={param.readOnly}
          aria-label={label}
          id={id}
          type="checkbox"
          onChange={onChange}
          checked={currentValue}
        />
        <CdsControlMessage className={isModified ? "italics" : ""}>
          {currentValue ? "true" : "false"}
        </CdsControlMessage>
      </CdsToggle>
      <CdsControlMessage>{unsavedMessage}</CdsControlMessage>
    </CdsToggleGroup>
  );

  return (
    <Row>
      <Column span={10}>{input}</Column>
    </Row>
  );
}
