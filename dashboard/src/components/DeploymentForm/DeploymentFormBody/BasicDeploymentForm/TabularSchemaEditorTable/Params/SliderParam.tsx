// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsControlMessage } from "@cds/react/forms";
import { CdsInput } from "@cds/react/input";
import { CdsRange } from "@cds/react/range";
import Column from "components/js/Column";
import Row from "components/js/Row";
import { toNumber } from "lodash";
import { useState } from "react";
import { IBasicFormParam } from "shared/types";
import { basicFormsDebounceTime } from "shared/utils";

export interface ISliderParamProps {
  id: string;
  label: string;
  param: IBasicFormParam;
  unit: string;
  step: number;
  handleBasicFormParamChange: (
    p: IBasicFormParam,
  ) => (e: React.FormEvent<HTMLInputElement>) => void;
}

export default function SliderParam(props: ISliderParamProps) {
  const { handleBasicFormParamChange, id, label, param, step } = props;

  const [currentValue, setCurrentValue] = useState(toNumber(param.currentValue) || param.minimum);
  const [isValueModified, setIsValueModified] = useState(false);
  const [timeout, setThisTimeout] = useState({} as NodeJS.Timeout);

  const onChange = (e: React.FormEvent<HTMLInputElement>) => {
    setCurrentValue(toNumber(e.currentTarget.value));
    setIsValueModified(toNumber(e.currentTarget.value) !== param.currentValue);
    // Gather changes before submitting
    clearTimeout(timeout);
    const func = handleBasicFormParamChange(param);
    // The reference to target get lost, so we need to keep a copy
    const targetCopy = {
      currentTarget: {
        value: e.currentTarget?.value,
        type: e.currentTarget?.type,
      },
    } as React.FormEvent<HTMLInputElement>;
    setThisTimeout(setTimeout(() => func(targetCopy), basicFormsDebounceTime));
  };

  const unsavedMessage = isValueModified ? "Unsaved" : "";
  const isModified =
    isValueModified ||
    (param.currentValue !== param.defaultValue && param.currentValue !== param.deployedValue);

  const input = (
    <CdsRange>
      <input
        aria-label={label}
        id={id + "_range"}
        type="range"
        min={Math.min(param.minimum || 0, currentValue)}
        max={Math.max(param.maximum || 1000, currentValue)}
        step={step}
        onChange={onChange}
        value={currentValue}
      />
      <CdsControlMessage>{unsavedMessage}</CdsControlMessage>
    </CdsRange>
  );

  const inputText = (
    <div className="self-center">
      <CdsInput className={isModified ? "bolder" : ""}>
        <input
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
