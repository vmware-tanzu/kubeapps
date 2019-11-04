import * as React from "react";
import Switch from "react-switch";
import { IBasicFormParam } from "shared/types";

export interface IStringParamProps {
  id: string;
  label: string;
  param: IBasicFormParam;
  handleBasicFormParamChange: (
    p: IBasicFormParam,
  ) => (e: React.FormEvent<HTMLInputElement>) => void;
}

class BooleanParam extends React.Component<IStringParamProps> {
  public render() {
    const { id, param, label } = this.props;
    return (
      <label htmlFor={id}>
        <div className="margin-b-normal">
          <span>{label}</span>
          <Switch
            id={id}
            onChange={this.handleChange}
            checked={!!param.value}
            className="react-switch"
          />
        </div>
        {param.description && <span className="description">{param.description}</span>}
      </label>
    );
  }

  // handleChange transform the event received by the Switch component to a checkbox event
  public handleChange = (checked: boolean) => {
    const { param } = this.props;
    const event = {
      currentTarget: { value: String(checked), type: "checkbox", checked },
    } as React.FormEvent<HTMLInputElement>;
    this.props.handleBasicFormParamChange(param)(event);
  };
}

export default BooleanParam;
