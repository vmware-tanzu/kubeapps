// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { useEffect, useState } from "react";
import { IBasicFormParam } from "shared/types";
import Slider from "../../Slider";

export interface ISliderParamProps {
  id: string;
  label: string;
  param: IBasicFormParam;
  unit: string;
  min: number;
  max: number;
  step: number;
  handleBasicFormParamChange: (
    p: IBasicFormParam,
  ) => (e: React.FormEvent<HTMLInputElement>) => void;
}

export interface ISliderParamState {
  value: number;
}

function toNumber(value: string | number) {
  // Force to return a Number from a string removing any character that is not a digit
  return typeof value === "number" ? value : Number(value.replace(/[^\d.]/g, ""));
}

function getDefaultValue(min: number, value?: string) {
  return (value && toNumber(value)) || min;
}

function SliderParam({
  id,
  label,
  param,
  unit,
  min,
  max,
  step,
  handleBasicFormParamChange,
}: ISliderParamProps) {
  const [value, setValue] = useState(getDefaultValue(min, param.value));

  useEffect(() => {
    setValue(getDefaultValue(min, param.value));
  }, [param, min]);

  const handleParamChange = (newValue: number) => {
    handleBasicFormParamChange(param)({
      currentTarget: {
        value: param.type === "string" ? `${newValue}${unit}` : newValue,
      },
    } as React.FormEvent<HTMLInputElement>);
  };

  // onChangeSlider is run when the slider is dropped at one point
  // at that point we update the parameter
  const onChangeSlider = (values: readonly number[]) => {
    handleParamChange(values[0]);
  };

  // onUpdateSlider is run when dragging the slider
  // we just update the state here for a faster response
  const onUpdateSlider = (values: readonly number[]) => {
    setValue(values[0]);
  };

  const onChangeInput = (e: React.FormEvent<HTMLInputElement>) => {
    const numberValue = toNumber(e.currentTarget.value);
    setValue(numberValue);
    handleParamChange(numberValue);
  };

  return (
    <div>
      <label htmlFor={id}>
        <span className="centered deployment-form-label deployment-form-label-text-param">
          {label}
        </span>
        <div className="slider-block">
          <div className="slider-content">
            <Slider
              // If the parameter defines a minimum or maximum, maintain those
              min={Math.min(param.minimum || min, value)}
              max={Math.max(param.maximum || max, value)}
              step={step || 1}
              default={value}
              onChange={onChangeSlider}
              onUpdate={onUpdateSlider}
              values={value}
              sliderStyle={{ width: "100%", margin: "1.2em 0 1.2em 0" }}
            />
          </div>
          <div className="slider-input-and-unit">
            <input
              className="slider-input clr-input"
              id={id}
              onChange={onChangeInput}
              value={value}
            />
            <span className="margin-l-normal">{unit}</span>
          </div>
        </div>
        {param.description && <span className="description">{param.description}</span>}
      </label>
    </div>
  );
}

export default SliderParam;
