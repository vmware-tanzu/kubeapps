// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { isArray } from "lodash";
import React from "react";
import { getValue } from "shared/schema";
import { DeploymentEvent, IBasicFormParam, IBasicFormSliderParam } from "shared/types";
import BooleanParam from "./BooleanParam";
import CustomFormComponentLoader from "./CustomFormParam";
import SliderParam from "./SliderParam";
import Subsection from "./Subsection";
import TextParam from "./TextParam";

interface IParamProps {
  appValues: string;
  param: IBasicFormParam;
  allParams: IBasicFormParam[];
  id: string;
  handleBasicFormParamChange: (
    p: IBasicFormParam,
  ) => (e: React.FormEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => void;
  handleValuesChange: (value: string) => void;
  deploymentEvent: DeploymentEvent;
}

export default function Param({
  appValues,
  param,
  allParams,
  id,
  handleBasicFormParamChange,
  handleValuesChange,
  deploymentEvent,
}: IParamProps) {
  let paramComponent: JSX.Element = <></>;

  const isHidden = () => {
    const hidden = param.hidden;
    switch (typeof hidden) {
      case "string":
        // If hidden is a string, it points to the value that should be true
        return evalCondition(hidden);
      case "object":
        // Two type of supported objects
        // A single condition: {value: string, path: any}
        // An array of conditions: {conditions: Array<{value: string, path: any}, operator: string}
        if (hidden.conditions?.length > 0) {
          // If hidden is an object, a different logic should be applied
          // based on the operator
          switch (hidden.operator) {
            case "and":
              // Every value matches the referenced
              // value (via jsonpath) in all the conditions
              return hidden.conditions.every(c => evalCondition(c.path, c.value, c.event));
            case "or":
              // It is enough if the value matches the referenced
              // value (via jsonpath) in any of the conditions
              return hidden.conditions.some(c => evalCondition(c.path, c.value, c.event));
            case "nor":
              // Every value mismatches the referenced
              // value (via jsonpath) in any of the conditions
              return hidden.conditions.every(c => !evalCondition(c.path, c.value, c.event));
            default:
              // we consider 'and' as the default operator
              return hidden.conditions.every(c => evalCondition(c.path, c.value, c.event));
          }
        } else {
          return evalCondition(hidden.path, hidden.value, hidden.event);
        }
      case "undefined":
        return false;
    }
  };

  const getParamMatchingPath = (params: IBasicFormParam[], path: string): any => {
    let targetParam;
    for (const p of params) {
      if (p.path === path) {
        targetParam = p;
        break;
      } else if (p.children && p.children?.length > 0) {
        targetParam = getParamMatchingPath(p.children, path);
      }
    }
    return targetParam;
  };

  const evalCondition = (
    path: string,
    expectedValue?: any,
    paramDeploymentEvent?: DeploymentEvent,
  ): boolean => {
    if (paramDeploymentEvent == null) {
      let val = getValue(appValues, path);
      // retrieve the value that the property pointed by path should have to be hidden.
      // https://github.com/vmware-tanzu/kubeapps/issues/1913
      if (val === undefined) {
        const target = getParamMatchingPath(allParams, path);
        val = target?.value;
      }
      return val === (expectedValue ?? true);
    } else {
      return paramDeploymentEvent === deploymentEvent;
    }
  };

  // Return early for custom components
  if (param.customComponent) {
    return (
      <div key={id} hidden={isHidden()}>
        <CustomFormComponentLoader
          param={param}
          handleBasicFormParamChange={handleBasicFormParamChange}
        />
      </div>
    );
  }

  // If the type of the param is an array, represent it as its first type
  const type = isArray(param.type) ? param.type[0] : param.type;
  if (type === "boolean") {
    paramComponent = (
      <BooleanParam
        label={param.title || param.path}
        handleBasicFormParamChange={handleBasicFormParamChange}
        id={id}
        param={param}
      />
    );
  } else if (type === "object") {
    paramComponent = (
      <Subsection
        label={param.title || param.path}
        handleValuesChange={handleValuesChange}
        appValues={appValues}
        param={param}
        allParams={allParams}
        deploymentEvent={deploymentEvent}
      />
    );
  } else if (param.render === "slider") {
    const p = param as IBasicFormSliderParam;
    paramComponent = (
      <SliderParam
        label={param.title || param.path}
        handleBasicFormParamChange={handleBasicFormParamChange}
        id={id}
        param={param}
        min={p.sliderMin || 1}
        max={p.sliderMax || 1000}
        step={p.sliderStep || 1}
        unit={p.sliderUnit || ""}
      />
    );
  } else if (param.render === "textArea") {
    paramComponent = (
      <TextParam
        label={param.title || param.path}
        handleBasicFormParamChange={handleBasicFormParamChange}
        id={id}
        param={param}
        inputType="textarea"
      />
    );
  } else {
    const label = param.title || param.path;
    let inputType = "string";
    if (type === "integer") {
      inputType = "number";
    }
    if (
      type === "string" &&
      (param.render === "password" || label.toLowerCase().includes("password"))
    ) {
      inputType = "password";
    }
    paramComponent = (
      <TextParam
        label={label}
        handleBasicFormParamChange={handleBasicFormParamChange}
        id={id}
        param={param}
        inputType={inputType}
      />
    );
  }

  return (
    <div key={id} hidden={isHidden()} className="basic-deployment-form-param">
      {paramComponent}
    </div>
  );
}
