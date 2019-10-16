import * as React from "react";
import { IBasicFormParam } from "shared/types";

export interface IStringParamProps {
  id: string;
  name: string;
  label: string;
  inputType?: string;
  param: IBasicFormParam;
  handleBasicFormParamChange: (
    name: string,
    p: IBasicFormParam,
  ) => (e: React.FormEvent<HTMLInputElement>) => void;
}

class TextParam extends React.Component<IStringParamProps> {
  public render() {
    const { id, name, param, label, inputType } = this.props;
    return (
      <div>
        <label htmlFor={id}>
          <div className="row">
            <div className="col-3 block">
              <div className="centered">{label}</div>
            </div>
            <div className="col-9 margin-t-small">
              <input
                id={id}
                onChange={this.props.handleBasicFormParamChange(name, param)}
                value={param.value === undefined ? "" : param.value}
                type={inputType ? inputType : "text"}
              />
              {param.description && <span className="description">{param.description}</span>}
            </div>
          </div>
        </label>
      </div>
    );
  }
}

export default TextParam;
