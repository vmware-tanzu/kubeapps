import * as React from "react";
import { setValue } from "../../../shared/schema";
import { IBasicFormParam } from "../../../shared/types";
import { getValueFromEvent } from "../../../shared/utils";

export interface ISubsectionProps {
  label: string;
  param: IBasicFormParam;
  handleValuesChange: (value: string) => void;
  renderParam: (
    param: IBasicFormParam,
    id: string,
    handleBasicFormParamChange: (
      p: IBasicFormParam,
    ) => (e: React.FormEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => void,
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
          param.children.map((childrenParam, i) => {
            return this.props.renderParam(
              childrenParam,
              `${childrenParam.path}-${i}`,
              this.handleChildrenParamChange,
            );
          })}
      </div>
    );
  }

  private handleChildrenParamChange = (param: IBasicFormParam) => {
    return (e: React.FormEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
      const value = getValueFromEvent(e);
      this.props.param.children = this.props.param.children!.map(p =>
        p.path === param.path ? { ...param, value } : p,
      );
      this.props.handleValuesChange(setValue(this.props.appValues, param.path, value));
    };
  };
}

export default Subsection;
