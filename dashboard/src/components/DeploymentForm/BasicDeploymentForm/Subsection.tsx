import * as React from "react";
import { setValue } from "../../../shared/schema";
import { IBasicFormParam } from "../../../shared/types";
import { getValueFromEvent } from "../../../shared/utils";
import BooleanParam from "./BooleanParam";
import TextParam from "./TextParam";

export interface ISubsectionProps {
  label: string;
  param: IBasicFormParam;
  name: string;
  enablerChildrenParam: string;
  enablerCondition: boolean;
  handleValuesChange: (value: string) => void;
  appValues: string;
}

class Subsection extends React.Component<ISubsectionProps> {
  public render() {
    const { label, param, name, enablerChildrenParam, enablerCondition } = this.props;
    return (
      <div className="subsection margin-v-normal">
        {param.children && param.children[enablerChildrenParam] && (
          <BooleanParam
            label={param.children[enablerChildrenParam].title || enablerChildrenParam}
            handleBasicFormParamChange={this.handleChildrenParamChange}
            id={`${name}-${enablerChildrenParam}`}
            name={enablerChildrenParam}
            param={param.children[enablerChildrenParam]}
          />
        )}
        <div
          hidden={
            param.children &&
            param.children[enablerChildrenParam] &&
            param.children[enablerChildrenParam].value !== enablerCondition
          }
        >
          <div className="margin-v-normal">{label}</div>
          {param.children &&
            Object.keys(param.children)
              .filter(p => p !== enablerChildrenParam)
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
      this.props.param.children![name] = { ...this.props.param.children![name], value };
      this.props.handleValuesChange(setValue(this.props.appValues, param.path, value));
    };
  };

  private renderParam(name: string, param: IBasicFormParam, index: number) {
    const id = `${name}-${index}`;
    switch (param.type) {
      case "boolean":
        return (
          <BooleanParam
            label={param.title || name}
            handleBasicFormParamChange={this.handleChildrenParamChange}
            key={id}
            id={id}
            name={name}
            param={param}
          />
        );
      default:
        return (
          <TextParam
            label={param.title || name}
            handleBasicFormParamChange={this.handleChildrenParamChange}
            key={id}
            id={id}
            name={name}
            param={param}
            inputType={param.type === "integer" ? "number" : "text"}
          />
        );
    }
  }
}

export default Subsection;
