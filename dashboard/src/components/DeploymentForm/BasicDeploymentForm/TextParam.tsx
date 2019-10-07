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
          {label}
          {param.description && (
            <>
              <br />
              <span className="description">{param.description}</span>
            </>
          )}
          <input
            id={id}
            onChange={this.props.handleBasicFormParamChange(name, param)}
            defaultValue={param.value}
            type={inputType ? inputType : "text"}
          />
        </label>
      </div>
    );
  }
}

export default TextParam;
