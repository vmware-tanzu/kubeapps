import * as React from "react";
import { IBasicFormParam } from "shared/types";

export interface IStringParamProps {
  id: string;
  label: string;
  param: IBasicFormParam;
  handleBasicFormParamChange: (
    p: IBasicFormParam,
  ) => (e: React.FormEvent<HTMLTextAreaElement>) => void;
}

class TextAreaParam extends React.Component<IStringParamProps> {
  public render() {
    const { id, param, label } = this.props;
    return (
      <div>
        <label htmlFor={id}>
          <div className="row">
            <div className="col-3 block">
              <div className="centered">{label}</div>
            </div>
            <div className="col-9 margin-t-small">
              <textarea
                id={id}
                onChange={this.props.handleBasicFormParamChange(param)}
                value={param.value === undefined ? "" : param.value}
              />
              {param.description && <span className="description">{param.description}</span>}
            </div>
          </div>
        </label>
      </div>
    );
  }
}

export default TextAreaParam;
