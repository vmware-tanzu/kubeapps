import * as React from "react";
import { IBasicFormParam } from "shared/types";
import Slider from "../../Slider";

export interface ISliderParamProps {
  id: string;
  label: string;
  param: IBasicFormParam;
  unit: string;
  min: number;
  max: number;
  step: number;
  handleBasicFormParamChange: (
    p: IBasicFormParam,
  ) => (e: React.FormEvent<HTMLInputElement>) => void;
}

export interface ISliderParamState {
  value: number;
}

function toNumber(value: string | number) {
  // Force to return a Number from a string removing any character that is not a digit
  return typeof value === "number" ? value : Number(value.replace(/[^\d.]/g, ""));
}

function getDefaultValue(min: number, value?: string) {
  return (value && toNumber(value)) || min;
}

class SliderParam extends React.Component<ISliderParamProps, ISliderParamState> {
  public state: ISliderParamState = {
    value: getDefaultValue(this.props.min, this.props.param.value),
  };

  public componentDidUpdate = (prevProps: ISliderParamProps) => {
    if (prevProps.param.value !== this.props.param.value) {
      const value = getDefaultValue(this.props.min, this.props.param.value);
      if (value !== this.state.value) {
        this.setState({ value });
      }
    }
  };

  // onChangeSlider is executed when the slider is dropped at one point
  // at that point we update the parameter
  public onChangeSlider = (values: readonly number[]) => {
    this.handleParamChange(values[0]);
  };

  // onUpdateSlider is executed when dragging the slider
  // we just update the state here for a faster response
  public onUpdateSlider = (values: readonly number[]) => {
    this.setState({ value: values[0] });
  };

  public onChangeInput = (e: React.FormEvent<HTMLInputElement>) => {
    const value = toNumber(e.currentTarget.value);
    this.setState({ value });
    this.handleParamChange(value);
  };

  public render() {
    const { param, label, min, max, step } = this.props;
    return (
      <div>
        <label htmlFor={this.props.id}>
          <div className="row">
            <div className="col-3 block">
              <div className="centered">{label}</div>
            </div>
            <div className="col-9 margin-t-small">
              <div className="row">
                <div className="col-9">
                  <Slider
                    // If the parameter defines a minimum or maximum, maintain those
                    min={Math.min(param.minimum || min, this.state.value)}
                    max={Math.max(param.maximum || max, this.state.value)}
                    step={step || 1}
                    default={this.state.value}
                    onChange={this.onChangeSlider}
                    onUpdate={this.onUpdateSlider}
                    values={this.state.value}
                  />
                </div>
                <div className="col-3">
                  <input
                    className="disk_size_input"
                    id={this.props.id}
                    onChange={this.onChangeInput}
                    value={this.state.value}
                  />
                  <span className="margin-l-normal">{this.props.unit}</span>
                </div>
              </div>
              {param.description && <span className="description">{param.description}</span>}
            </div>
          </div>
        </label>
      </div>
    );
  }

  private handleParamChange = (value: number) => {
    this.props.handleBasicFormParamChange(this.props.param)({
      currentTarget: {
        value: this.props.param.type === "string" ? `${value}${this.props.unit}` : value,
      },
    } as React.FormEvent<HTMLInputElement>);
  };
}

export default SliderParam;
