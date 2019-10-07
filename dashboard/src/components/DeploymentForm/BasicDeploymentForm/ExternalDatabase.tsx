import * as React from "react";
import { IBasicFormParam } from "shared/types";
import BooleanParam from "./BooleanParam";
import TextParam from "./TextParam";

export interface IStringParamProps {
  label: string;
  externalDatabaseParams: { [name: string]: IBasicFormParam };
  handleBasicFormParamChange: (
    name: string,
    p: IBasicFormParam,
  ) => (e: React.FormEvent<HTMLInputElement>) => void;
}

// These are the keys that should be used in the `schema` to be recognized
export const USE_SELF_HOSTED_DB_PARAM_NAME = "useSelfHostedDatabase";
export const EXTERNAL_DB_PARAM_NAME = "externalDatabase";
export const EXTERNAL_DB_HOST_PARAM_NAME = "externalDatabaseHost";
export const EXTERNAL_DB_USER_PARAM_NAME = "externalDatabaseUser";
export const EXTERNAL_DB_PASSWORD_PARAM_NAME = "externalDatabasePassword";
export const EXTERNAL_DB_NAME_PARAM_NAME = "externalDatabaseDB";
export const EXTERNAL_DB_PORT_PARAM_NAME = "externalDatabasePort";

class ExternalDatabaseSection extends React.Component<IStringParamProps> {
  public render() {
    const { label, externalDatabaseParams } = this.props;
    const selfHostedDatabaseParam = externalDatabaseParams.useSelfHostedDatabase;
    return (
      <div className="subsection margin-v-normal">
        <BooleanParam
          label="Use a Self Hosted Database"
          handleBasicFormParamChange={this.props.handleBasicFormParamChange}
          id={"enable-self-hosted-db"}
          name={USE_SELF_HOSTED_DB_PARAM_NAME}
          param={selfHostedDatabaseParam}
        />
        <div hidden={selfHostedDatabaseParam.value} className="margin-t-normal">
          <span>{label}</span>
          {Object.keys(externalDatabaseParams).map((paramName, i) => {
            return this.renderParam(paramName, externalDatabaseParams[paramName], i);
          })}
        </div>
      </div>
    );
  }
  private renderParam(name: string, param: IBasicFormParam, index: number) {
    const id = `${name}-${index}`;
    let label = "";
    let type = "text";
    switch (name) {
      case EXTERNAL_DB_HOST_PARAM_NAME:
        label = "Host";
        break;
      case EXTERNAL_DB_USER_PARAM_NAME:
        label = "User";
        break;
      case EXTERNAL_DB_PASSWORD_PARAM_NAME:
        label = "Password";
        break;
      case EXTERNAL_DB_NAME_PARAM_NAME:
        label = "Database";
        break;
      case EXTERNAL_DB_PORT_PARAM_NAME:
        label = "Port";
        type = "number";
        break;
      case USE_SELF_HOSTED_DB_PARAM_NAME:
        // Handled in the main render function
        return;
    }
    return (
      <TextParam
        label={label}
        handleBasicFormParamChange={this.props.handleBasicFormParamChange}
        key={id}
        id={id}
        name={name}
        param={param}
        inputType={type}
      />
    );
  }
}

export default ExternalDatabaseSection;
