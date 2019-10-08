import * as React from "react";
import { IBasicFormParam } from "shared/types";
import BooleanParam from "./BooleanParam";
import TextParam from "./TextParam";

export interface IDatabaseSectionProps {
  label: string;
  param: IBasicFormParam;
  disableExternalDBParamName: string;
  disableExternalDBParam: IBasicFormParam;
  handleBasicFormParamChange: (
    name: string,
    p: IBasicFormParam,
  ) => (e: React.FormEvent<HTMLInputElement>) => void;
}

class DatabaseSection extends React.Component<IDatabaseSectionProps> {
  public render() {
    const { label, param, disableExternalDBParam, disableExternalDBParamName } = this.props;
    return (
      <div className="subsection margin-v-normal">
        <BooleanParam
          label="Use a Self Hosted Database"
          handleBasicFormParamChange={this.props.handleBasicFormParamChange}
          id={"enable-self-hosted-db"}
          name={disableExternalDBParamName}
          param={disableExternalDBParam}
        />
        <div hidden={disableExternalDBParam.value} className="margin-t-normal">
          <span>{label}</span>
          {param.children &&
            Object.keys(param.children).map((paramName, i) => {
              return this.renderParam(paramName, param.children![paramName], i);
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
      case "externalDatabaseHost":
        label = "Host";
        break;
      case "externalDatabaseUser":
        label = "User";
        break;
      case "externalDatabasePassword":
        label = "Password";
        break;
      case "externalDatabaseName":
        label = "Database";
        break;
      case "externalDatabasePort":
        label = "Port";
        type = "number";
        break;
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

export default DatabaseSection;
