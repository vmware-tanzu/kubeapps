import * as React from "react";
import { setValue } from "../../../shared/schema";
import { IBasicFormParam } from "../../../shared/types";
import { getValueFromEvent } from "../../../shared/utils";
import BooleanParam from "./BooleanParam";
import TextParam from "./TextParam";

export interface IDatabaseSectionProps {
  label: string;
  param: IBasicFormParam;
  disableExternalDBParamName: string;
  disableExternalDBParam: IBasicFormParam;
  handleValuesChange: (value: string) => void;
  appValues: string;
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

  private handleChildrenParamChange = (name: string, param: IBasicFormParam) => {
    return (e: React.FormEvent<HTMLInputElement>) => {
      const value = getValueFromEvent(e);
      this.props.handleValuesChange(setValue(this.props.appValues, param.path, value));
      this.props.param.children![name] = { ...this.props.param.children![name], value };
    };
  };

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
        handleBasicFormParamChange={this.handleChildrenParamChange}
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
