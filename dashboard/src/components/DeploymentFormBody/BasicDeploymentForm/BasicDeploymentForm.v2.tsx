import * as React from "react";
import { DeploymentEvent, IBasicFormParam } from "shared/types";

import "./BasicDeploymentForm.v2.css";
import Param from "./Param";

export interface IBasicDeploymentFormProps {
  deploymentEvent: DeploymentEvent;
  params: IBasicFormParam[];
  handleBasicFormParamChange: (
    p: IBasicFormParam,
  ) => (e: React.FormEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => void;
  handleValuesChange: (value: string) => void;
  appValues: string;
}

function BasicDeploymentForm(props: IBasicDeploymentFormProps) {
  return (
    <div className="deployment-form-tabs-data">
      {props.params.map((param, i) => {
        const id = `${param.path}-${i}`;
        return (
          <div key={id}>
            <Param {...props} param={param} id={id} />
            <hr className="param-separator" />
          </div>
        );
      })}
    </div>
  );
}

export default BasicDeploymentForm;
