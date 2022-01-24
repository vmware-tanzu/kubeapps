// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { setValue } from "shared/schema";
import { DeploymentEvent, IBasicFormParam } from "shared/types";
import { getValueFromEvent } from "shared/utils";
import Param from "./Param";

export interface ISubsectionProps {
  label: string;
  param: IBasicFormParam;
  allParams: IBasicFormParam[];
  appValues: string;
  deploymentEvent: DeploymentEvent;
  handleValuesChange: (value: string) => void;
}

function Subsection({
  label,
  param,
  allParams,
  appValues,
  deploymentEvent,
  handleValuesChange,
}: ISubsectionProps) {
  const handleChildrenParamChange = (childrenParam: IBasicFormParam) => {
    return (e: React.FormEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
      const value = getValueFromEvent(e);
      param.children = param.children!.map(p =>
        p.path === childrenParam.path ? { ...childrenParam, value } : p,
      );
      handleValuesChange(setValue(appValues, childrenParam.path, value));
    };
  };

  return (
    <div className="subsection">
      <div>
        <label className="deployment-form-label">{label}</label>
        {param.description && (
          <>
            <br />
            <span className="description">{param.description}</span>
          </>
        )}
      </div>
      {param.children &&
        param.children.map((childrenParam, i) => {
          const id = `${childrenParam.path}-${i}`;
          return (
            <Param
              param={childrenParam}
              allParams={allParams}
              id={id}
              key={id}
              handleBasicFormParamChange={handleChildrenParamChange}
              deploymentEvent={deploymentEvent}
              appValues={appValues}
              handleValuesChange={handleValuesChange}
            />
          );
        })}
    </div>
  );
}

export default Subsection;
