import { shallow } from "enzyme";
import * as React from "react";
import { IBasicFormParam } from "shared/types";
import Slider from "../../Slider";
import SliderParam from "./SliderParam";

const defaultProps = {
  id: "disk",
  label: "Disk Size",
  handleBasicFormParamChange: jest.fn(() => jest.fn()),
  min: 1,
  max: 100,
  unit: "Gi",
};

const params = [
  {
    value: "10Gi",
    type: "string",
    path: "disk",
  } as IBasicFormParam,
  {
    value: 10,
    type: "integer",
    path: "disk",
  } as IBasicFormParam,
  {
    value: 10.0,
    type: "number",
    path: "disk",
  } as IBasicFormParam,
];

it("renders a disk size param with a default value", () => {
  params.forEach(param => {
    const wrapper = shallow(<SliderParam {...defaultProps} param={param} />);
    expect(wrapper.state("value")).toBe(10);
    expect(wrapper).toMatchSnapshot();
  });
});

it("changes the value of the string param when the slider changes", () => {
  params.forEach(param => {
    const cloneParam = { ...param } as IBasicFormParam;
    const expected = param.type === "string" ? "20Gi" : 20;

    const handleBasicFormParamChange = jest.fn(() => {
      cloneParam.value = expected;
      return jest.fn();
    });

    const wrapper = shallow(
      <SliderParam
        {...defaultProps}
        param={cloneParam}
        handleBasicFormParamChange={handleBasicFormParamChange}
      />,
    );

    expect(wrapper.state("value")).toBe(10);

    const slider = wrapper.find(Slider);
    (slider.prop("onChange") as (values: number[]) => void)([20]);

    expect(cloneParam.value).toBe(expected);
    expect(handleBasicFormParamChange.mock.calls[0]).toEqual([
      { value: expected, type: param.type, path: param.path },
    ]);
  });
});

it("updates state but does not change param value during slider update (only when dropped in a point)", () => {
  params.forEach(param => {
    const handleBasicFormParamChange = jest.fn();
    const wrapper = shallow(
      <SliderParam
        {...defaultProps}
        param={param}
        handleBasicFormParamChange={handleBasicFormParamChange}
      />,
    );
    expect(wrapper.state("value")).toBe(10);

    const slider = wrapper.find(Slider);
    (slider.prop("onUpdate") as (values: number[]) => void)([20]);

    expect(wrapper.state("value")).toBe(20);
    expect(handleBasicFormParamChange).not.toHaveBeenCalled();
  });
});

describe("when changing the value in the input", () => {
  it("parses a number and forwards it", () => {
    params.forEach(param => {
      const valueChange = jest.fn();
      const handleBasicFormParamChange = jest.fn(() => valueChange);
      const wrapper = shallow(
        <SliderParam
          {...defaultProps}
          param={param}
          handleBasicFormParamChange={handleBasicFormParamChange}
        />,
      );
      expect(wrapper.state("value")).toBe(10);

      const input = wrapper.find("input#disk");
      const event = { currentTarget: { value: "20" } } as React.FormEvent<HTMLInputElement>;
      (input.prop("onChange") as (e: React.FormEvent<HTMLInputElement>) => void)(event);

      expect(wrapper.state("value")).toBe(20);

      const expected = param.type === "string" ? "20Gi" : 20;
      expect(valueChange.mock.calls[0]).toEqual([{ currentTarget: { value: expected } }]);
    });
  });

  it("ignores values in the input that are not digits", () => {
    params.forEach(param => {
      const valueChange = jest.fn();
      const handleBasicFormParamChange = jest.fn(() => valueChange);
      const wrapper = shallow(
        <SliderParam
          {...defaultProps}
          param={param}
          handleBasicFormParamChange={handleBasicFormParamChange}
        />,
      );
      expect(wrapper.state("value")).toBe(10);

      const input = wrapper.find("input#disk");
      const event = { currentTarget: { value: "foo20*#@$" } } as React.FormEvent<HTMLInputElement>;
      (input.prop("onChange") as (e: React.FormEvent<HTMLInputElement>) => void)(event);

      expect(wrapper.state("value")).toBe(20);

      const expected = param.type === "string" ? "20Gi" : 20;
      expect(valueChange.mock.calls[0]).toEqual([{ currentTarget: { value: expected } }]);
    });
  });

  it("accept decimal values", () => {
    params.forEach(param => {
      const valueChange = jest.fn();
      const handleBasicFormParamChange = jest.fn(() => valueChange);
      const wrapper = shallow(
        <SliderParam
          {...defaultProps}
          param={param}
          handleBasicFormParamChange={handleBasicFormParamChange}
        />,
      );
      expect(wrapper.state("value")).toBe(10);

      const input = wrapper.find("input#disk");
      const event = { currentTarget: { value: "20.5" } } as React.FormEvent<HTMLInputElement>;
      (input.prop("onChange") as (e: React.FormEvent<HTMLInputElement>) => void)(event);

      expect(wrapper.state("value")).toBe(20.5);

      const expected = param.type === "string" ? "20.5Gi" : 20.5;
      expect(valueChange.mock.calls[0]).toEqual([{ currentTarget: { value: expected } }]);
    });
  });

  it("modifies the max value of the slider if the input is bigger than 100", () => {
    params.forEach(param => {
      const valueChange = jest.fn();
      const handleBasicFormParamChange = jest.fn(() => valueChange);
      const wrapper = shallow(
        <SliderParam
          {...defaultProps}
          param={param}
          handleBasicFormParamChange={handleBasicFormParamChange}
        />,
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
  params.forEach(param => {
    const cloneParam = { ...param } as IBasicFormParam;
    cloneParam.value = undefined;

    const wrapper = shallow(<SliderParam {...defaultProps} param={cloneParam} min={5} />);
    expect(wrapper.state("value")).toBe(5);
  });
});

it("updates the state when receiving new props", () => {
  params.forEach(param => {
    const handleBasicFormParamChange = jest.fn();
    const wrapper = shallow(
      <SliderParam
        {...defaultProps}
        param={param}
        handleBasicFormParamChange={handleBasicFormParamChange}
      />,
    );
    expect(wrapper.state("value")).toBe(10);

    const newValue = param.type === "string" ? "20Gi" : 20;
    wrapper.setProps({ param: { value: newValue } });
    expect(wrapper.state("value")).toBe(20);
    expect(handleBasicFormParamChange).not.toHaveBeenCalled();
  });
});
