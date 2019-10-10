import * as React from "react";
import { setValue, USE_SELF_HOSTED_DB } from "../../../shared/schema";
import { IBasicFormParam } from "../../../shared/types";
import { getValueFromEvent } from "../../../shared/utils";
import BooleanParam from "./BooleanParam";
import TextParam from "./TextParam";

export interface IDatabaseSectionProps {
  label: string;
  param: IBasicFormParam;
  handleValuesChange: (value: string) => void;
  appValues: string;
}

class DatabaseSection extends React.Component<IDatabaseSectionProps> {
  public render() {
    const { label, param } = this.props;
    return (
      <div className="subsection margin-v-normal">
        <BooleanParam
          label="Use a Self Hosted Database"
          handleBasicFormParamChange={this.handleChildrenParamChange}
          id={"enable-self-hosted-db"}
          name={USE_SELF_HOSTED_DB}
          param={param.children![USE_SELF_HOSTED_DB]}
        />
        <div hidden={param.children![USE_SELF_HOSTED_DB].value} className="margin-t-normal">
          <span>{label}</span>
          {param.children &&
            Object.keys(param.children)
              .filter(p => p !== USE_SELF_HOSTED_DB)
              .map((paramName, i) => {
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
