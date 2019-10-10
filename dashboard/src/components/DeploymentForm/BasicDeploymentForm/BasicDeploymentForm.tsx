import * as React from "react";
import { IBasicFormParam } from "shared/types";
import TextParam from "./TextParam";

import { DISK_SIZE, EXTERNAL_DB, USE_SELF_HOSTED_DB } from "../../../shared/schema";
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
    return Object.keys(this.props.params).map((paramName, i) => {
      return this.renderParam(
        paramName,
        this.props.params[paramName],
        i,
        this.props.handleBasicFormParamChange,
      );
    });
  }

  private renderParam(
    name: string,
    param: IBasicFormParam,
    index: number,
    handleBasicFormParamChange: (
      name: string,
      p: IBasicFormParam,
    ) => (e: React.FormEvent<HTMLInputElement>) => void,
  ) {
    const id = `${name}-${index}`;
    switch (name) {
      case EXTERNAL_DB:
        return (
          <Subsection
            label={param.title || "External Database Details"}
            handleValuesChange={this.props.handleValuesChange}
            appValues={this.props.appValues}
            renderParam={this.renderParam}
            key={id}
            name={name}
            param={param}
            enablerChildrenParam={USE_SELF_HOSTED_DB}
            enablerCondition={false}
          />
        );
      case DISK_SIZE:
        return (
          <SliderParam
            label={param.title || "Disk Size"}
            handleBasicFormParamChange={handleBasicFormParamChange}
            key={id}
            id={id}
            name={name}
            param={param}
            min={1}
            max={100}
            unit="Gi"
          />
        );
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
