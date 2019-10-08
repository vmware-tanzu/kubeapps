import * as React from "react";
import { IBasicFormParam } from "shared/types";
import StringParam from "./StringParam";

import "./BasicDeploymentForm.css";

export interface IBasicDeploymentFormProps {
  params: { [name: string]: IBasicFormParam };
  handleBasicFormParamChange: (
    name: string,
    p: IBasicFormParam,
  ) => (e: React.FormEvent<HTMLInputElement>) => void;
}

class BasicDeploymentForm extends React.Component<IBasicDeploymentFormProps> {
  public render() {
    return Object.keys(this.props.params).map((paramName, i) => {
      return this.renderParam(paramName, this.props.params[paramName], i);
    });
  }

  private renderParam(name: string, param: IBasicFormParam, index: number) {
    const id = `${name}-${index}`;
    switch (name) {
      case "username":
        return (
          <StringParam
            label="Username"
            handleBasicFormParamChange={this.props.handleBasicFormParamChange}
            key={id}
            id={id}
            name={name}
            param={param}
          />
        );
      case "password":
        return (
          <StringParam
            label="Password"
            handleBasicFormParamChange={this.props.handleBasicFormParamChange}
            key={id}
            id={id}
            name={name}
            param={param}
          />
        );
      case "email":
        return (
          <StringParam
            label="Email"
            handleBasicFormParamChange={this.props.handleBasicFormParamChange}
            key={id}
            id={id}
            name={name}
            param={param}
          />
        );
      default:
        if (param.type === "string") {
          return (
            <StringParam
              label={param.title || ""}
              handleBasicFormParamChange={this.props.handleBasicFormParamChange}
              key={id}
              id={id}
              name={name}
              param={param}
            />
          );
        }
      // TODO(andres): This should return an error once we add support for all the parameters that we expect
      // throw new Error(`Param ${name} with type ${param.type} is not supported`);
    }
    return null;
  }
}

export default BasicDeploymentForm;
