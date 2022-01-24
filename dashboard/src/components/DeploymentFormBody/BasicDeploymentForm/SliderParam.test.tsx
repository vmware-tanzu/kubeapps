// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { shallow } from "enzyme";
import React from "react";
import { IBasicFormParam } from "shared/types";
import Slider from "../../Slider";
import SliderParam from "./SliderParam";

const defaultProps = {
  id: "disk",
  label: "Disk Size",
  handleBasicFormParamChange: jest.fn(() => jest.fn()),
  min: 1,
  max: 100,
  step: 1,
  unit: "Gi",
};

const params: IBasicFormParam[] = [
  {
    value: "10Gi",
    type: "string",
    path: "disk",
  },
  {
    value: 10,
    type: "integer",
    path: "disk",
  },
  {
    value: 10.0,
    type: "number",
    path: "disk",
  },
];

it("renders a disk size param with a default value", () => {
  params.forEach(param => {
    const wrapper = shallow(<SliderParam {...defaultProps} param={param} />);
    expect(wrapper.find(Slider).prop("values")).toBe(10);
    expect(wrapper).toMatchSnapshot();
  });
});

describe("when changing the slide", () => {
  it("changes the value of the string param", () => {
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

      expect(wrapper.find(Slider).prop("values")).toBe(10);

      const slider = wrapper.find(Slider);
      (slider.prop("onChange") as (values: number[]) => void)([20]);

      expect(cloneParam.value).toBe(expected);
      expect(handleBasicFormParamChange.mock.calls[0]).toEqual([
        { value: expected, type: param.type, path: param.path },
      ]);
    });
  });

  it("changes the value of the string param without unit", () => {
    params.forEach(param => {
      const cloneParam = { ...param } as IBasicFormParam;
      const expected = param.type === "string" ? "20" : 20;

      const handleBasicFormParamChange = jest.fn(() => {
        cloneParam.value = expected;
        return jest.fn();
      });

      const wrapper = shallow(
        <SliderParam
          {...defaultProps}
          param={cloneParam}
          unit=""
          handleBasicFormParamChange={handleBasicFormParamChange}
        />,
      );

      expect(wrapper.find(Slider).prop("values")).toBe(10);

      const slider = wrapper.find(Slider);
      (slider.prop("onChange") as (values: number[]) => void)([20]);

      expect(cloneParam.value).toBe(expected);
      expect(handleBasicFormParamChange.mock.calls[0]).toEqual([
        { value: expected, type: param.type, path: param.path },
      ]);
    });
  });

  it("changes the value of the string param with the step defined", () => {
    params.forEach(param => {
      const cloneProps = { ...defaultProps, step: 10 };
      const cloneParam = { ...param } as IBasicFormParam;
      const expected = param.type === "string" ? "20Gi" : 20;

      const handleBasicFormParamChange = jest.fn(() => {
        cloneParam.value = expected;
        return jest.fn();
      });

      const wrapper = shallow(
        <SliderParam
          {...cloneProps}
          param={cloneParam}
          handleBasicFormParamChange={handleBasicFormParamChange}
        />,
      );

      expect(wrapper.find(Slider).prop("values")).toBe(10);

      const slider = wrapper.find(Slider);
      (slider.prop("onChange") as (values: number[]) => void)([2]);

      expect(cloneParam.value).toBe(expected);
      expect(handleBasicFormParamChange.mock.calls[0]).toEqual([
        { value: expected, type: param.type, path: param.path },
      ]);
    });
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
    expect(wrapper.find(Slider).prop("values")).toBe(10);

    const slider = wrapper.find(Slider);
    (slider.prop("onUpdate") as (values: number[]) => void)([20]);

    expect(wrapper.find(Slider).prop("values")).toBe(20);
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
      expect(wrapper.find(Slider).prop("values")).toBe(10);

      const input = wrapper.find("input#disk");
      const event = { currentTarget: { value: "20" } } as React.FormEvent<HTMLInputElement>;
      (input.prop("onChange") as (e: React.FormEvent<HTMLInputElement>) => void)(event);

      expect(wrapper.find(Slider).prop("values")).toBe(20);

      const expected = param.type === "string" ? "20Gi" : 20;
      expect(valueChange.mock.calls[0]).toEqual([{ currentTarget: { value: expected } }]);
    });
  });

  it("parses a number and forwards it without unit", () => {
    params.forEach(param => {
      const valueChange = jest.fn();
      const handleBasicFormParamChange = jest.fn(() => valueChange);
      const wrapper = shallow(
        <SliderParam
          {...defaultProps}
          unit=""
          param={param}
          handleBasicFormParamChange={handleBasicFormParamChange}
        />,
      );
      expect(wrapper.find(Slider).prop("values")).toBe(10);

      const input = wrapper.find("input#disk");
      const event = { currentTarget: { value: "20" } } as React.FormEvent<HTMLInputElement>;
      (input.prop("onChange") as (e: React.FormEvent<HTMLInputElement>) => void)(event);

      expect(wrapper.find(Slider).prop("values")).toBe(20);

      const expected = param.type === "string" ? "20" : 20;
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
      expect(wrapper.find(Slider).prop("values")).toBe(10);

      const input = wrapper.find("input#disk");
      const event = { currentTarget: { value: "foo20*#@$" } } as React.FormEvent<HTMLInputElement>;
      (input.prop("onChange") as (e: React.FormEvent<HTMLInputElement>) => void)(event);

      expect(wrapper.find(Slider).prop("values")).toBe(20);

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
      expect(wrapper.find(Slider).prop("values")).toBe(10);

      const input = wrapper.find("input#disk");
      const event = { currentTarget: { value: "20.5" } } as React.FormEvent<HTMLInputElement>;
      (input.prop("onChange") as (e: React.FormEvent<HTMLInputElement>) => void)(event);

      expect(wrapper.find(Slider).prop("values")).toBe(20.5);

      const expected = param.type === "string" ? "20.5Gi" : 20.5;
      expect(valueChange.mock.calls[0]).toEqual([{ currentTarget: { value: expected } }]);
    });
  });

  it("modifies the max value of the slider if the input is greater than 100", () => {
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
      expect(wrapper.find(Slider).prop("values")).toBe(10);

      const input = wrapper.find("input#disk");
      const event = { currentTarget: { value: "200" } } as React.FormEvent<HTMLInputElement>;
      (input.prop("onChange") as (e: React.FormEvent<HTMLInputElement>) => void)(event);

      expect(wrapper.find(Slider).prop("values")).toBe(200);
      const slider = wrapper.find(Slider);
      expect(slider.prop("max")).toBe(200);
    });
  });
});

it("uses the param minimum and maximum if defined", () => {
  params.forEach(param => {
    const clonedParam = { ...param } as IBasicFormParam;
    clonedParam.minimum = 5;
    clonedParam.maximum = 50;

    const wrapper = shallow(<SliderParam {...defaultProps} param={clonedParam} />);

    const slider = wrapper.find(Slider);
    expect(slider.prop("min")).toBe(5);
    expect(slider.prop("max")).toBe(50);
  });
});

it("defaults to the min if the value is undefined", () => {
  params.forEach(param => {
    const cloneParam = { ...param } as IBasicFormParam;
    cloneParam.value = undefined;

    const wrapper = shallow(<SliderParam {...defaultProps} param={cloneParam} min={5} />);
    expect(wrapper.find(Slider).prop("values")).toBe(5);
  });
});
