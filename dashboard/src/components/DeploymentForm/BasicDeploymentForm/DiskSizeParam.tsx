import * as React from "react";
import { IBasicFormParam } from "shared/types";
import Slider from "../../../components/Slider";

export interface IDiskSizeParamProps {
  id: string;
  name: string;
  label: string;
  param: IBasicFormParam;
  handleBasicFormParamChange: (
    name: string,
    p: IBasicFormParam,
  ) => (e: React.FormEvent<HTMLInputElement>) => void;
}

export interface IDiskSizeParamState {
  Gi: number;
}

function toNumber(value: string) {
  // Force to return a Number from a string removing any character that is not a digit
  return Number(value.replace(/[^\d\.]/g, ""));
}

class DiskSizeParam extends React.Component<IDiskSizeParamProps, IDiskSizeParamState> {
  public state: IDiskSizeParamState = {
    Gi: toNumber(this.props.param.value) || 10,
  };

  // onChangeSlider is executed when the slider is dropped at one point
  // at that point we update the parameter
  public onChangeSlider = (values: number[]) => {
    this.handleParamChange(values[0]);
  };

  // onUpdateSlider is executed when dragging the slider
  // we just update the state here for a faster response
  public onUpdateSlider = (values: number[]) => {
    this.setState({ Gi: values[0] });
  };

  public onChangeInput = (e: React.FormEvent<HTMLInputElement>) => {
    const value = toNumber(e.currentTarget.value);
    this.setState({ Gi: value });
    this.handleParamChange(value);
  };

  public render() {
    const { param, label } = this.props;
    return (
      <div>
        <label htmlFor={this.props.id}>
          {label}
          {param.description && (
            <>
              <br />
              <span className="description">{param.description}</span>
            </>
          )}
          <div className="row">
            <div className="col-10">
              <Slider
                min={1}
                max={Math.max(100, this.state.Gi)}
                default={this.state.Gi}
                onChange={this.onChangeSlider}
                onUpdate={this.onUpdateSlider}
                values={this.state.Gi}
              />
            </div>
            <div className="col-2">
              <input
                className="disk_size_input"
                id={this.props.id}
                onChange={this.onChangeInput}
                value={this.state.Gi}
              />
              <span className="margin-l-normal">Gi</span>
            </div>
          </div>
        </label>
      </div>
    );
  }

  private handleParamChange = (value: number) => {
    this.props.handleBasicFormParamChange(this.props.name, this.props.param)({
      currentTarget: { value: `${value}Gi` },
    } as React.FormEvent<HTMLInputElement>);
  };
}

export default DiskSizeParam;
