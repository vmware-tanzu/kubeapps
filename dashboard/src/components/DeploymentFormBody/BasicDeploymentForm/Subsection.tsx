import * as React from "react";
import { setValue } from "../../../shared/schema";
import { IBasicFormParam } from "../../../shared/types";
import { getValueFromEvent } from "../../../shared/utils";

export interface ISubsectionProps {
  label: string;
  param: IBasicFormParam;
  name: string;
  handleValuesChange: (value: string) => void;
  renderParam: (
    name: string,
    param: IBasicFormParam,
    id: string,
    handleBasicFormParamChange: (
      name: string,
      p: IBasicFormParam,
    ) => (e: React.FormEvent<HTMLInputElement>) => void,
  ) => JSX.Element | null;
  appValues: string;
}

class Subsection extends React.Component<ISubsectionProps> {
  public render() {
    const { label, param } = this.props;
    return (
      <div className="subsection margin-v-normal">
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
          Object.keys(param.children).map((paramName, i) => {
            return this.props.renderParam(
              paramName,
              param.children![paramName],
              `${paramName}-${i}`,
              this.handleChildrenParamChange,
            );
          })}
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
