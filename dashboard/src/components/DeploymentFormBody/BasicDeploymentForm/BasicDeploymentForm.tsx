import { isArray } from "lodash";
import * as React from "react";
import { IBasicFormParam, IBasicFormSliderParam } from "shared/types";
import TextParam from "./TextParam";

import { getValue } from "../../../shared/schema";
import "./BasicDeploymentForm.css";
import BooleanParam from "./BooleanParam";
import SliderParam from "./SliderParam";
import Subsection from "./Subsection";

export interface IBasicDeploymentFormProps {
  params: IBasicFormParam[];
  handleBasicFormParamChange: (
    p: IBasicFormParam,
  ) => (e: React.FormEvent<HTMLInputElement | HTMLTextAreaElement>) => void;
  handleValuesChange: (value: string) => void;
  appValues: string;
}

class BasicDeploymentForm extends React.Component<IBasicDeploymentFormProps> {
  public render() {
    return (
      <div className="margin-t-normal">
        {this.props.params.map((param, i) => {
          const id = `${param.path}-${i}`;
          return (
            <div key={id}>
              {this.renderParam(param, id, this.props.handleBasicFormParamChange)}
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
        let isHiddenParam;
        // If hidden is an object, a different logic should be applied
        // based on the operator
        switch (hidden.operator) {
          case "and":
            // Every value matches its reference in
            // all the conditions
            isHiddenParam = true;
            hidden.conditions.forEach(c => {
              if (getValue(this.props.appValues, c.value) !== c.reference) {
                isHiddenParam = false;
              }
            });
            return isHiddenParam;
          case "or":
            // It is enough if the value matches its reference in
            // any of the conditions
            isHiddenParam = false;
            hidden.conditions.forEach(c => {
              if (getValue(this.props.appValues, c.value) === c.reference) {
                isHiddenParam = true;
              }
            });
            return isHiddenParam;
          default:
            return false;
        }
      case "undefined":
        return false;
    }
  };

  private renderParam = (
    param: IBasicFormParam,
    id: string,
    handleBasicFormParamChange: (
      p: IBasicFormParam,
    ) => (e: React.FormEvent<HTMLInputElement | HTMLTextAreaElement>) => void,
  ) => {
    let paramComponent: JSX.Element = <></>;
    // If the type of the param is an array, represent it as its first type
    const type = isArray(param.type) ? param.type[0] : param.type;
    switch (type) {
      case "boolean":
        paramComponent = (
          <BooleanParam
            label={param.title || param.path}
            handleBasicFormParamChange={handleBasicFormParamChange}
            id={id}
            param={param}
          />
        );
        break;
      case "object": {
        paramComponent = (
          <Subsection
            label={param.title || param.path}
            handleValuesChange={this.props.handleValuesChange}
            appValues={this.props.appValues}
            renderParam={this.renderParam}
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
              label={param.title || param.path}
              handleBasicFormParamChange={handleBasicFormParamChange}
              id={id}
              param={param}
              min={p.sliderMin || 1}
              max={p.sliderMax || 1000}
              unit={p.sliderUnit || ""}
            />
          );
          break;
        }
        if (param.render === "textArea") {
          paramComponent = (
            <TextParam
              label={param.title || name}
              handleBasicFormParamChange={handleBasicFormParamChange}
              id={id}
              param={param}
              inputType="textarea"
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
            param={param}
            inputType={type === "integer" ? "number" : "string"}
          />
        );
    }
    return (
      <div key={id} hidden={this.isHidden(param)}>
        {paramComponent}
      </div>
    );
  };
}

export default BasicDeploymentForm;
