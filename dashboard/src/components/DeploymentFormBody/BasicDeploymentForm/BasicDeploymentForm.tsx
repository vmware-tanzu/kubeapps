import * as React from "react";
import { IBasicFormParam } from "shared/types";
import TextParam from "./TextParam";

import {
  CPU_REQUEST,
  DISK_SIZE,
  ENABLE_INGRESS,
  EXTERNAL_DB,
  INGRESS,
  MEMORY_REQUEST,
  RESOURCES,
  USE_SELF_HOSTED_DB,
} from "../../../shared/schema";
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
      case RESOURCES:
        return (
          <Subsection
            label={param.title || "Application resources"}
            handleValuesChange={this.props.handleValuesChange}
            appValues={this.props.appValues}
            renderParam={this.renderParam}
            key={id}
            name={name}
            param={param}
          />
        );
      case MEMORY_REQUEST:
        return (
          <SliderParam
            label={param.title || "Memory Request"}
            handleBasicFormParamChange={handleBasicFormParamChange}
            key={id}
            id={id}
            name={name}
            param={param}
            min={10}
            max={2048}
            unit="Mi"
          />
        );
      case CPU_REQUEST:
        return (
          <SliderParam
            label={param.title || "CPU Request"}
            handleBasicFormParamChange={handleBasicFormParamChange}
            key={id}
            id={id}
            name={name}
            param={param}
            min={10}
            max={2000}
            unit="m"
          />
        );
      case INGRESS:
        return (
          <Subsection
            label={param.title || "Ingress details"}
            handleValuesChange={this.props.handleValuesChange}
            appValues={this.props.appValues}
            renderParam={this.renderParam}
            key={id}
            name={name}
            param={param}
            enablerChildrenParam={ENABLE_INGRESS}
            enablerCondition={true}
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
