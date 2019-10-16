import * as React from "react";
import { setValue } from "../../../shared/schema";
import { IBasicFormParam } from "../../../shared/types";
import { getValueFromEvent } from "../../../shared/utils";
import BooleanParam from "./BooleanParam";

export interface ISubsectionProps {
  label: string;
  param: IBasicFormParam;
  name: string;
  handleValuesChange: (value: string) => void;
  renderParam: (
    name: string,
    param: IBasicFormParam,
    index: number,
    handleBasicFormParamChange: (
      name: string,
      p: IBasicFormParam,
    ) => (e: React.FormEvent<HTMLInputElement>) => void,
  ) => JSX.Element | null;
  appValues: string;
}

export interface ISubsectionState {
  enablerChildrenParam: string;
  enablerCondition: boolean;
}

function findEnabler(name: string, param: IBasicFormParam) {
  let result = { enablerChildrenParam: "", enablerCondition: false };
  const children = param.children;
  if (children) {
    Object.keys(children).forEach(p => {
      if (children[p].type === "boolean") {
        if (children[p].enables === name) {
          result = { enablerChildrenParam: p, enablerCondition: true };
        } else if (children[p].disables === name) {
          result = { enablerChildrenParam: p, enablerCondition: false };
        }
      }
    });
  }
  return result;
}

class Subsection extends React.Component<ISubsectionProps> {
  public state: ISubsectionState = findEnabler(this.props.name, this.props.param);

  public render() {
    const { label, param, name } = this.props;
    const { enablerChildrenParam, enablerCondition } = this.state;
    return (
      <div className="subsection margin-v-normal">
        {param.children && enablerChildrenParam && param.children[enablerChildrenParam] && (
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
            !!enablerChildrenParam &&
            param.children[enablerChildrenParam] &&
            param.children[enablerChildrenParam].value !== enablerCondition
          }
        >
          <div className="margin-v-normal">
            {label}
            {param.description && (
              <>
                <br />
                <span className="description">{param.description}</span>
              </>
            )}
          </div>

          {param.children &&
            Object.keys(param.children)
              .filter(p => p !== enablerChildrenParam)
              .map((paramName, i) => {
                return this.props.renderParam(
                  paramName,
                  param.children![paramName],
                  i,
                  this.handleChildrenParamChange,
                );
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
}

export default Subsection;
