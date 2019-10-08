import * as React from "react";
import { IBasicFormParam } from "shared/types";
import TextParam from "./TextParam";

import "./BasicDeploymentForm.css";
import BooleanParam from "./BooleanParam";
import DatabaseSection from "./DatabaseSection";
import DiskSizeParam from "./DiskSizeParam";

export interface IBasicDeploymentFormProps {
  params: { [name: string]: IBasicFormParam };
  handleBasicFormParamChange: (
    name: string,
    p: IBasicFormParam,
  ) => (e: React.FormEvent<HTMLInputElement>) => void;
}

const USE_SELF_HOSTED_DB_PARAM_NAME = "useSelfHostedDatabase";

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
          <TextParam
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
          <TextParam
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
          <TextParam
            label="Email"
            handleBasicFormParamChange={this.props.handleBasicFormParamChange}
            key={id}
            id={id}
            name={name}
            param={param}
          />
        );
      case "externalDatabase":
        return (
          <DatabaseSection
            label="External Database Details"
            handleBasicFormParamChange={this.props.handleBasicFormParamChange}
            key={id}
            param={param}
            disableExternalDBParamName={USE_SELF_HOSTED_DB_PARAM_NAME}
            disableExternalDBParam={this.props.params[USE_SELF_HOSTED_DB_PARAM_NAME]}
          />
        );
      case USE_SELF_HOSTED_DB_PARAM_NAME:
        // Handled within ExternalDabataseSection
        break;
      case "diskSize":
        return (
          <DiskSizeParam
            label="Disk Size"
            handleBasicFormParamChange={this.props.handleBasicFormParamChange}
            key={id}
            id={id}
            name={name}
            param={param}
          />
        );
      default:
        switch (param.type) {
          case "string":
            return (
              <TextParam
                label={param.title || ""}
                handleBasicFormParamChange={this.props.handleBasicFormParamChange}
                key={id}
                id={id}
                name={name}
                param={param}
              />
            );
          case "integer":
            return (
              <TextParam
                label={param.title || ""}
                handleBasicFormParamChange={this.props.handleBasicFormParamChange}
                key={id}
                id={id}
                name={name}
                param={param}
                inputType="number"
              />
            );
          case "boolean":
            return (
              <BooleanParam
                label={param.title || ""}
                handleBasicFormParamChange={this.props.handleBasicFormParamChange}
                key={id}
                id={id}
                name={name}
                param={param}
              />
            );
          default:
          // TODO(andres): This should return an error once we add support for all the parameters that we expect
          // throw new Error(`Param ${name} with type ${param.type} is not supported`);
        }
    }
    return null;
  }
}

export default BasicDeploymentForm;
