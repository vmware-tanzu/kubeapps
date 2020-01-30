import * as React from "react";
import { IBasicFormParam } from "shared/types";

export interface IStringParamProps {
  id: string;
  label: string;
  inputType?: string;
  param: IBasicFormParam;
  handleBasicFormParamChange: (
    p: IBasicFormParam,
  ) => (e: React.FormEvent<HTMLInputElement | HTMLTextAreaElement>) => void;
}

class TextParam extends React.Component<IStringParamProps> {
  public render() {
    const { id, param, label, inputType } = this.props;
    let input = (
      <input
        id={id}
        onChange={this.props.handleBasicFormParamChange(param)}
        value={param.value === undefined ? "" : param.value}
        type={inputType ? inputType : "text"}
      />
    );
    if (inputType === "textarea") {
      input = (
        <textarea
          id={id}
          onChange={this.props.handleBasicFormParamChange(param)}
          value={param.value === undefined ? "" : param.value}
        />
      );
    }
    return (
      <div>
        <label htmlFor={id}>
          <div className="row">
            <div className="col-3 block">
              <div className="centered">{label}</div>
            </div>
            <div className="col-9 margin-t-small">
              {input}
              {param.description && <span className="description">{param.description}</span>}
            </div>
          </div>
        </label>
      </div>
    );
  }
}

export default TextParam;
