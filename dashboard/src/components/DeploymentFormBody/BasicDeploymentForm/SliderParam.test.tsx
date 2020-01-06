import { shallow } from "enzyme";
import * as React from "react";
import { IBasicFormParam } from "shared/types";
import Slider from "../../Slider";
import SliderParam from "./SliderParam";

const defaultProps = {
  id: "disk",
  label: "Disk Size",
  param: {
    value: "10Gi",
    type: "string",
    path: "disk",
  } as IBasicFormParam,
  handleBasicFormParamChange: jest.fn(() => jest.fn()),
  min: 1,
  max: 100,
  unit: "Gi",
};

it("renders a disk size param with a default value", () => {
  const wrapper = shallow(<SliderParam {...defaultProps} />);
  expect(wrapper.state("value")).toBe(10);
  expect(wrapper).toMatchSnapshot();
});

it("changes the value of the param when the slider changes", () => {
  const param = {
    value: "10Gi",
    type: "string",
    path: "disk",
  } as IBasicFormParam;
  const handleBasicFormParamChange = jest.fn(() => {
    param.value = "20Gi";
    return jest.fn();
  });

  const wrapper = shallow(
    <SliderParam
      {...defaultProps}
      param={param}
      handleBasicFormParamChange={handleBasicFormParamChange}
    />,
  );
  expect(wrapper.state("value")).toBe(10);

  const slider = wrapper.find(Slider);
  (slider.prop("onChange") as (values: number[]) => void)([20]);

  expect(param.value).toBe("20Gi");
  expect(handleBasicFormParamChange.mock.calls[0]).toEqual([
    { value: "20Gi", type: "string", path: "disk" },
  ]);
});

it("updates state but does not change param value during slider update (only when dropped in a point)", () => {
  const handleBasicFormParamChange = jest.fn();
  const wrapper = shallow(
    <SliderParam {...defaultProps} handleBasicFormParamChange={handleBasicFormParamChange} />,
  );
  expect(wrapper.state("value")).toBe(10);

  const slider = wrapper.find(Slider);
  (slider.prop("onUpdate") as (values: number[]) => void)([20]);

  expect(wrapper.state("value")).toBe(20);
  expect(handleBasicFormParamChange).not.toHaveBeenCalled();
});

describe("when changing the value in the input", () => {
  it("parses a number and forwards it", () => {
    const valueChange = jest.fn();
    const handleBasicFormParamChange = jest.fn(() => valueChange);
    const wrapper = shallow(
      <SliderParam {...defaultProps} handleBasicFormParamChange={handleBasicFormParamChange} />,
    );
    expect(wrapper.state("value")).toBe(10);

    const input = wrapper.find("input#disk");
    const event = { currentTarget: { value: "20" } } as React.FormEvent<HTMLInputElement>;
    (input.prop("onChange") as (e: React.FormEvent<HTMLInputElement>) => void)(event);

    expect(wrapper.state("value")).toBe(20);
    expect(valueChange.mock.calls[0]).toEqual([{ currentTarget: { value: "20Gi" } }]);
  });

  it("ignores values in the input that are not digits", () => {
    const valueChange = jest.fn();
    const handleBasicFormParamChange = jest.fn(() => valueChange);
    const wrapper = shallow(
      <SliderParam {...defaultProps} handleBasicFormParamChange={handleBasicFormParamChange} />,
    );
    expect(wrapper.state("value")).toBe(10);

    const input = wrapper.find("input#disk");
    const event = { currentTarget: { value: "foo20*#@$" } } as React.FormEvent<HTMLInputElement>;
    (input.prop("onChange") as (e: React.FormEvent<HTMLInputElement>) => void)(event);

    expect(wrapper.state("value")).toBe(20);
    expect(valueChange.mock.calls[0]).toEqual([{ currentTarget: { value: "20Gi" } }]);
  });

  it("accept decimal values", () => {
    const valueChange = jest.fn();
    const handleBasicFormParamChange = jest.fn(() => valueChange);
    const wrapper = shallow(
      <SliderParam {...defaultProps} handleBasicFormParamChange={handleBasicFormParamChange} />,
    );
    expect(wrapper.state("value")).toBe(10);

    const input = wrapper.find("input#disk");
    const event = { currentTarget: { value: "20.5" } } as React.FormEvent<HTMLInputElement>;
    (input.prop("onChange") as (e: React.FormEvent<HTMLInputElement>) => void)(event);

    expect(wrapper.state("value")).toBe(20.5);
    expect(valueChange.mock.calls[0]).toEqual([{ currentTarget: { value: "20.5Gi" } }]);
  });

  it("modifies the max value of the slider if the input is bigger than 100", () => {
    const valueChange = jest.fn();
    const handleBasicFormParamChange = jest.fn(() => valueChange);
    const wrapper = shallow(
      <SliderParam {...defaultProps} handleBasicFormParamChange={handleBasicFormParamChange} />,
    );
    expect(wrapper.state("value")).toBe(10);

    const input = wrapper.find("input#disk");
    const event = { currentTarget: { value: "200" } } as React.FormEvent<HTMLInputElement>;
    (input.prop("onChange") as (e: React.FormEvent<HTMLInputElement>) => void)(event);

    expect(wrapper.state("value")).toBe(200);
    const slider = wrapper.find(Slider);
    expect(slider.prop("max")).toBe(200);
  });
});

it("uses the param minimum and maximum if defined", () => {
  const param = {
    value: "10Gi",
    type: "string",
    path: "disk",
    minimum: 5,
    maximum: 50,
  } as IBasicFormParam;

  const wrapper = shallow(<SliderParam {...defaultProps} param={param} />);

  const slider = wrapper.find(Slider);
  expect(slider.prop("min")).toBe(5);
  expect(slider.prop("max")).toBe(50);
});

it("defaults to the min if the value is undefined", () => {
  const param = {
    type: "string",
    path: "disk",
  } as IBasicFormParam;

  const wrapper = shallow(<SliderParam {...defaultProps} param={param} min={5} />);

  expect(wrapper.state("value")).toBe(5);
});

it("updates the state when receiving new props", () => {
  const handleBasicFormParamChange = jest.fn();
  const wrapper = shallow(
    <SliderParam {...defaultProps} handleBasicFormParamChange={handleBasicFormParamChange} />,
  );
  expect(wrapper.state("value")).toBe(10);

  wrapper.setProps({ param: { value: "20Gi" } });
  expect(wrapper.state("value")).toBe(20);
  expect(handleBasicFormParamChange).not.toHaveBeenCalled();
});
