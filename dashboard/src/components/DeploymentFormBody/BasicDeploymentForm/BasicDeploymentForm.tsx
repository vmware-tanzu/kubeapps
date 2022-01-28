// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { DeploymentEvent, IBasicFormParam } from "shared/types";
import "./BasicDeploymentForm.css";
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
            <Param {...props} allParams={props.params} param={param} id={id} />
            <hr className="param-separator" />
          </div>
        );
      })}
    </div>
  );
}

export default BasicDeploymentForm;
