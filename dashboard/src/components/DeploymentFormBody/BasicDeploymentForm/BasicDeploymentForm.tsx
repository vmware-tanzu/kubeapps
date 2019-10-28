import * as React from "react";
import { IBasicFormParam, IBasicFormSliderParam } from "shared/types";
import TextParam from "./TextParam";

import { getValue } from "../../../shared/schema";
import "./BasicDeploymentForm.css";
import BooleanParam from "./BooleanParam";
import SliderParam from "./SliderParam";
import Subsection from "./Subsection";

export interface IBasicDeploymentFormProps {
  params: { [name: string]: IBasicFormParam };
  handleBasicFormParamChange: (
    name: string,
    p: IBasicFormParam,
  ) => (e: React.FormEvent<HTMLInputElement>) => void;
  handleValuesChange: (value: string) => void;
  appValues: string;
}

class BasicDeploymentForm extends React.Component<IBasicDeploymentFormProps> {
  public render() {
    return (
      <div className="margin-t-normal">
        {Object.keys(this.props.params).map((paramName, i) => {
          const id = `${paramName}-${i}`;
          return (
            <div key={id}>
              {this.renderParam(
                paramName,
                this.props.params[paramName],
                id,
                this.props.handleBasicFormParamChange,
              )}
              <hr />
            </div>
          );
        })}
      </div>
    );
  }

  private isHidden = (param: IBasicFormParam) => {
    const hidden = param.hidden;
    switch (typeof hidden) {
      case "string":
        // If hidden is a string, it points to the value that should be true
        return getValue(this.props.appValues, hidden) === true;
      case "object":
        // If hidden is an object, inspect the value it points to
        return getValue(this.props.appValues, hidden.value) === hidden.condition;
      case "undefined":
        return false;
    }
  };

  private renderParam = (
    name: string,
    param: IBasicFormParam,
    id: string,
    handleBasicFormParamChange: (
      name: string,
      p: IBasicFormParam,
    ) => (e: React.FormEvent<HTMLInputElement>) => void,
  ) => {
    let paramComponent: JSX.Element = <></>;
    switch (name) {
      default:
        switch (param.type) {
          case "boolean":
            paramComponent = (
              <BooleanParam
                label={param.title || name}
                handleBasicFormParamChange={handleBasicFormParamChange}
                id={id}
                name={name}
                param={param}
              />
            );
            break;
          case "object": {
            paramComponent = (
              <Subsection
                label={param.title || name}
                handleValuesChange={this.props.handleValuesChange}
                appValues={this.props.appValues}
                renderParam={this.renderParam}
                name={name}
                param={param}
              />
            );
            break;
          }
          case "string": {
            if (param.render === "slider") {
              const p = param as IBasicFormSliderParam;
              paramComponent = (
                <SliderParam
                  label={param.title || name}
                  handleBasicFormParamChange={handleBasicFormParamChange}
                  id={id}
                  name={name}
                  param={param}
                  min={p.sliderMin || 1}
                  max={p.sliderMax || 1000}
                  unit={p.sliderUnit || ""}
                />
              );
              break;
            }
          }
          default:
            paramComponent = (
              <TextParam
                label={param.title || name}
                handleBasicFormParamChange={handleBasicFormParamChange}
                id={id}
                name={name}
                param={param}
                inputType={param.type === "integer" ? "number" : "string"}
              />
            );
        }
    }
    return (
      <div key={id} hidden={this.isHidden(param)}>
        {paramComponent}
      </div>
    );
  };
}

export default BasicDeploymentForm;
