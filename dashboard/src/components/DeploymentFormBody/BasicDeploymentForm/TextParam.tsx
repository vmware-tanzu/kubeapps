import * as React from "react";
import { IBasicFormParam } from "shared/types";

export interface IStringParamProps {
  id: string;
  label: string;
  inputType?: string;
  param: IBasicFormParam;
  handleBasicFormParamChange: (
    param: IBasicFormParam,
  ) => (e: React.FormEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => void;
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
    } else if (param.enum != null && param.enum.length > 0) {
      input = (
        <select id={id} onChange={this.props.handleBasicFormParamChange(param)}>
          {param.enum.map(enumValue => (
            <option key={enumValue} selected={param.value === enumValue}>
              {enumValue}
            </option>
          ))}
        </select>
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
