import * as React from "react";
import { IBasicFormParam } from "shared/types";

export interface IStringParamProps {
  id: string;
  name: string;
  label: string;
  param: IBasicFormParam;
  handleBasicFormParamChange: (
    name: string,
    p: IBasicFormParam,
  ) => (e: React.FormEvent<HTMLInputElement>) => void;
}

class StringParam extends React.Component<IStringParamProps> {
  public render() {
    const { id, name, param, label } = this.props;
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
            value={param.value}
          />
        </label>
      </div>
    );
  }
}

export default StringParam;
