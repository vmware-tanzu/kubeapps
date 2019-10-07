import * as React from "react";
import { IBasicFormParam } from "shared/types";
import TextParam from "./TextParam";

import "./BasicDeploymentForm.css";
import BooleanParam from "./BooleanParam";
import ExternalDatabaseSection, {
  EXTERNAL_DB_HOST_PARAM_NAME,
  EXTERNAL_DB_NAME_PARAM_NAME,
  EXTERNAL_DB_PARAM_NAME,
  EXTERNAL_DB_PASSWORD_PARAM_NAME,
  EXTERNAL_DB_PORT_PARAM_NAME,
  EXTERNAL_DB_USER_PARAM_NAME,
  USE_SELF_HOSTED_DB_PARAM_NAME,
} from "./ExternalDatabase";

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
      case EXTERNAL_DB_PARAM_NAME:
        return (
          <ExternalDatabaseSection
            label="External Database Details"
            handleBasicFormParamChange={this.props.handleBasicFormParamChange}
            key={id}
            externalDatabaseParams={this.filterDatabaseParams()}
          />
        );
      case EXTERNAL_DB_HOST_PARAM_NAME:
      case EXTERNAL_DB_USER_PARAM_NAME:
      case EXTERNAL_DB_PASSWORD_PARAM_NAME:
      case EXTERNAL_DB_NAME_PARAM_NAME:
      case EXTERNAL_DB_PORT_PARAM_NAME:
      case USE_SELF_HOSTED_DB_PARAM_NAME:
        // Handled within ExternalDabataseSection
        break;
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

  private filterDatabaseParams() {
    let databaseParams = {};
    Object.keys(this.props.params).map(paramName => {
      switch (paramName) {
        case EXTERNAL_DB_HOST_PARAM_NAME:
        case EXTERNAL_DB_USER_PARAM_NAME:
        case EXTERNAL_DB_PASSWORD_PARAM_NAME:
        case EXTERNAL_DB_NAME_PARAM_NAME:
        case EXTERNAL_DB_PORT_PARAM_NAME:
        case USE_SELF_HOSTED_DB_PARAM_NAME:
          databaseParams = {
            ...databaseParams,
            [paramName]: this.props.params[paramName],
          };
      }
    });
    return databaseParams;
  }
}

export default BasicDeploymentForm;
