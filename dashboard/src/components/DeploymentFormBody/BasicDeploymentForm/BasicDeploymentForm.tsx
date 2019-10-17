import * as React from "react";
import { IBasicFormParam, IBasicFormSliderParam } from "shared/types";
import TextParam from "./TextParam";

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

  private renderParam(
    name: string,
    param: IBasicFormParam,
    id: string,
    handleBasicFormParamChange: (
      name: string,
      p: IBasicFormParam,
    ) => (e: React.FormEvent<HTMLInputElement>) => void,
  ) {
    switch (name) {
      default:
        switch (param.type) {
          case "boolean":
            return (
              <BooleanParam
                label={param.title || name}
                handleBasicFormParamChange={handleBasicFormParamChange}
                key={id}
                id={id}
                name={name}
                param={param}
              />
            );
          case "object": {
            return (
              <Subsection
                label={param.title || name}
                handleValuesChange={this.props.handleValuesChange}
                appValues={this.props.appValues}
                renderParam={this.renderParam}
                key={id}
                name={name}
                param={param}
              />
            );
          }
          case "string": {
            if (param.render === "slider") {
              const p = param as IBasicFormSliderParam;
              return (
                <SliderParam
                  label={param.title || name}
                  handleBasicFormParamChange={handleBasicFormParamChange}
                  key={id}
                  id={id}
                  name={name}
                  param={param}
                  min={p.sliderMin || 1}
                  max={p.sliderMax || 1000}
                  unit={p.sliderUnit || ""}
                />
              );
            }
          }
          default:
            return (
              <TextParam
                label={param.title || name}
                handleBasicFormParamChange={handleBasicFormParamChange}
                key={id}
                id={id}
                name={name}
                param={param}
                inputType={param.type === "integer" ? "number" : "string"}
              />
            );
        }
    }
    return null;
  }
}

export default BasicDeploymentForm;
