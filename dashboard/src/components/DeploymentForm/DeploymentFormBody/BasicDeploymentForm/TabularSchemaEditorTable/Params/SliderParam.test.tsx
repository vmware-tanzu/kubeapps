// Copyright 2019-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { act } from "@testing-library/react";
import { shallow } from "enzyme";
import { IBasicFormParam } from "shared/types";
import SliderParam, { ISliderParamProps } from "./SliderParam";

const defaultProps = {
  id: "disk",
  label: "Disk Size",
  handleBasicFormParamChange: jest.fn(() => jest.fn()),
  step: 1,
  unit: "Gi",
  param: {} as IBasicFormParam,
} as ISliderParamProps;

const params = [
  {
    key: "disk",
    path: "disk",
    type: "integer",
    title: "Disk Size",
    hasProperties: false,
    isRequired: false,
    currentValue: 10,
    defaultValue: 10,
    deployedValue: 10,
    schema: {
      type: "integer",
    },
  },
  {
    key: "disk",
    path: "disk",
    type: "number",
    title: "Disk Size",
    hasProperties: false,
    isRequired: false,
    currentValue: 10.0,
    defaultValue: 10.0,
    deployedValue: 10.0,
    schema: {
      type: "number",
    },
  },
] as IBasicFormParam[];

jest.useFakeTimers();

it("renders a disk size param with a default value", () => {
  params.forEach(param => {
    const wrapper = shallow(<SliderParam {...defaultProps} param={param} />);
    const slider = wrapper.find("input").filterWhere(i => i.prop("type") === "range");
    expect(slider.prop("value")).toBe(10);

    const input = wrapper.find("input").filterWhere(i => i.prop("type") === "number");
    expect(input.prop("value")).toBe(10);
  });
});

it("uses the param minimum and maximum if defined", () => {
  params.forEach(param => {
    const wrapper = shallow(
      <SliderParam
        {...defaultProps}
        param={{ ...param, currentValue: 5, minimum: 5, maximum: 50 }}
      />,
    );

    const slider = wrapper.find("input").filterWhere(i => i.prop("type") === "range");
    expect(slider.prop("value")).toBe(5);
    expect(slider.prop("min")).toBe(5);
    expect(slider.prop("max")).toBe(50);

    const input = wrapper.find("input").filterWhere(i => i.prop("type") === "number");
    expect(input.prop("value")).toBe(5);
    expect(input.prop("min")).toBe(5);
    expect(input.prop("max")).toBe(50);
  });
});

it("uses the param exclusiveMinimum and exclusiveMaximum if defined", () => {
  params.forEach(param => {
    const wrapper = shallow(
      <SliderParam
        {...defaultProps}
        param={{ ...param, currentValue: 5, exclusiveMinimum: 5, exclusiveMaximum: 50 }}
      />,
    );

    const slider = wrapper.find("input").filterWhere(i => i.prop("type") === "range");
    expect(slider.prop("value")).toBe(5);
    expect(slider.prop("min")).toBe(5);
    expect(slider.prop("max")).toBe(50);

    const input = wrapper.find("input").filterWhere(i => i.prop("type") === "number");
    expect(input.prop("value")).toBe(5);
    expect(input.prop("min")).toBe(5);
    expect(input.prop("max")).toBe(50);
  });
});

it("uses the param the lowest minimum/exclusiveMinimum and maximum/exclusiveMaximum if defined", () => {
  params.forEach(param => {
    const wrapper = shallow(
      <SliderParam
        {...defaultProps}
        param={{
          ...param,
          currentValue: 5,
          minimum: 7,
          exclusiveMinimum: 5,
          maximum: 55,
          exclusiveMaximum: 50,
        }}
      />,
    );

    const slider = wrapper.find("input").filterWhere(i => i.prop("type") === "range");
    expect(slider.prop("value")).toBe(5);
    expect(slider.prop("min")).toBe(5);
    expect(slider.prop("max")).toBe(50);

    const input = wrapper.find("input").filterWhere(i => i.prop("type") === "number");
    expect(input.prop("value")).toBe(5);
    expect(input.prop("min")).toBe(5);
    expect(input.prop("max")).toBe(50);
  });
});

it("does not set the param minimum to current value if less than min", () => {
  params.forEach(param => {
    const wrapper = shallow(
      <SliderParam
        {...defaultProps}
        param={{ ...param, currentValue: 1, exclusiveMinimum: 100, maximum: 100 }}
      />,
    );

    const slider = wrapper.find("input").filterWhere(i => i.prop("type") === "range");
    expect(slider.prop("value")).toBe(1);
    expect(slider.prop("min")).toBe(100);
    expect(slider.prop("max")).toBe(100);

    const input = wrapper.find("input").filterWhere(i => i.prop("type") === "number");
    expect(input.prop("value")).toBe(1);
    expect(input.prop("min")).toBe(100);
    expect(input.prop("max")).toBe(100);
  });
});

it("does not set the param maximum to current value if greater than", () => {
  params.forEach(param => {
    const wrapper = shallow(
      <SliderParam
        {...defaultProps}
        param={{ ...param, currentValue: 2000, minimum: 100, maximum: 100 }}
      />,
    );

    const slider = wrapper.find("input").filterWhere(i => i.prop("type") === "range");
    expect(slider.prop("value")).toBe(2000);
    expect(slider.prop("min")).toBe(100);
    expect(slider.prop("max")).toBe(100);

    const input = wrapper.find("input").filterWhere(i => i.prop("type") === "number");
    expect(input.prop("value")).toBe(2000);
    expect(input.prop("min")).toBe(100);
    expect(input.prop("max")).toBe(100);
  });
});

it("defaults to the min if the value is undefined", () => {
  params.forEach(param => {
    const wrapper = shallow(
      <SliderParam {...defaultProps} param={{ ...param, currentValue: undefined, minimum: 5 }} />,
    );
    const slider = wrapper.find("input").filterWhere(i => i.prop("type") === "range");
    expect(slider.prop("value")).toBe(5);

    const input = wrapper.find("input").filterWhere(i => i.prop("type") === "number");
    expect(input.prop("value")).toBe(5);
  });
});

it("add required property if the param is required", () => {
  params.forEach(param => {
    const wrapper = shallow(
      <SliderParam {...defaultProps} param={{ ...param, isRequired: true }} />,
    );
    const slider = wrapper.find("input").filterWhere(i => i.prop("type") === "range");
    expect(slider.prop("required")).toBe(true);

    const input = wrapper.find("input").filterWhere(i => i.prop("type") === "number");
    expect(input.prop("required")).toBe(true);
  });
});

it("add disabled property if the param is readOnly", () => {
  params.forEach(param => {
    const wrapper = shallow(<SliderParam {...defaultProps} param={{ ...param, readOnly: true }} />);
    const slider = wrapper.find("input").filterWhere(i => i.prop("type") === "range");
    expect(slider.prop("disabled")).toBe(true);

    const input = wrapper.find("input").filterWhere(i => i.prop("type") === "number");
    expect(input.prop("disabled")).toBe(true);
  });
});

describe("when changing the slide", () => {
  it("changes the value of the string param", () => {
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
      expect(
        wrapper
          .find("input")
          .filterWhere(i => i.prop("type") === "range")
          .prop("value"),
      ).toBe(10);

      const event = {
        currentTarget: { value: "20", reportValidity: jest.fn() },
      } as unknown as React.FormEvent<HTMLInputElement>;
      act(() => {
        (
          wrapper.find("input#disk_range").prop("onChange") as (
            e: React.FormEvent<HTMLInputElement>,
          ) => void
        )(event);
      });
      wrapper.update();
      jest.runAllTimers();

      expect(
        wrapper
          .find("input")
          .filterWhere(i => i.prop("type") === "number")
          .prop("value"),
      ).toBe(20);

      expect(
        wrapper
          .find("input")
          .filterWhere(i => i.prop("type") === "range")
          .prop("value"),
      ).toBe(20);

      expect(handleBasicFormParamChange).toHaveBeenCalledWith(param);
      expect(valueChange).toHaveBeenCalledWith({
        ...event,
        currentTarget: { ...event.currentTarget, reportValidity: undefined },
      });
    });
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
      const slider = wrapper.find("input").filterWhere(i => i.prop("type") === "range");
      expect(slider.prop("value")).toBe(10);

      const input = wrapper.find("input").filterWhere(i => i.prop("type") === "number");
      expect(input.prop("value")).toBe(10);

      const event = {
        currentTarget: { value: "20", reportValidity: jest.fn() },
      } as unknown as React.FormEvent<HTMLInputElement>;
      act(() => {
        (
          wrapper.find("input#disk_text").prop("onChange") as (
            e: React.FormEvent<HTMLInputElement>,
          ) => void
        )(event);
      });
      wrapper.update();
      jest.runAllTimers();

      const sliderChanged = wrapper.find("input").filterWhere(i => i.prop("type") === "range");
      expect(sliderChanged.prop("value")).toBe(20);

      const inputChanged = wrapper.find("input").filterWhere(i => i.prop("type") === "number");
      expect(inputChanged.prop("value")).toBe(20);

      expect(handleBasicFormParamChange).toHaveBeenCalledWith(param);
      expect(valueChange).toHaveBeenCalledWith({
        ...event,
        currentTarget: { ...event.currentTarget, reportValidity: undefined },
      });
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
      const slider = wrapper.find("input").filterWhere(i => i.prop("type") === "range");
      expect(slider.prop("value")).toBe(10);

      const input = wrapper.find("input").filterWhere(i => i.prop("type") === "number");
      expect(input.prop("value")).toBe(10);

      const event = {
        currentTarget: { value: "foo20*#@$", reportValidity: jest.fn() },
      } as unknown as React.FormEvent<HTMLInputElement>;
      act(() => {
        (
          wrapper.find("input#disk_text").prop("onChange") as (
            e: React.FormEvent<HTMLInputElement>,
          ) => void
        )(event);
      });
      wrapper.update();
      jest.runAllTimers();

      const sliderChanged = wrapper.find("input").filterWhere(i => i.prop("type") === "range");
      expect(sliderChanged.prop("value")).toBe(NaN);

      const inputChanged = wrapper.find("input").filterWhere(i => i.prop("type") === "number");
      expect(inputChanged.prop("value")).toBe(NaN);

      expect(handleBasicFormParamChange).toHaveBeenCalledWith(param);
      expect(valueChange).toHaveBeenCalledWith({
        ...event,
        currentTarget: { ...event.currentTarget, reportValidity: undefined },
      });
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
      const slider = wrapper.find("input").filterWhere(i => i.prop("type") === "range");
      expect(slider.prop("value")).toBe(10);

      const input = wrapper.find("input").filterWhere(i => i.prop("type") === "number");
      expect(input.prop("value")).toBe(10);

      const event = {
        currentTarget: { value: "20.5", reportValidity: jest.fn() },
      } as unknown as React.FormEvent<HTMLInputElement>;
      act(() => {
        (
          wrapper.find("input#disk_text").prop("onChange") as (
            e: React.FormEvent<HTMLInputElement>,
          ) => void
        )(event);
      });
      wrapper.update();
      jest.runAllTimers();

      const sliderChanged = wrapper.find("input").filterWhere(i => i.prop("type") === "range");
      expect(sliderChanged.prop("value")).toBe(20.5);

      const inputChanged = wrapper.find("input").filterWhere(i => i.prop("type") === "number");
      expect(inputChanged.prop("value")).toBe(20.5);

      expect(handleBasicFormParamChange).toHaveBeenCalledWith(param);
      expect(valueChange).toHaveBeenCalledWith({
        ...event,
        currentTarget: { ...event.currentTarget, reportValidity: undefined },
      });
    });
  });

  it("does not modify the max value of the slider if the input is greater than min", () => {
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
      const slider = wrapper.find("input").filterWhere(i => i.prop("type") === "range");
      expect(slider.prop("value")).toBe(10);

      const input = wrapper.find("input").filterWhere(i => i.prop("type") === "number");
      expect(input.prop("value")).toBe(10);

      const event = {
        currentTarget: { value: "2000", reportValidity: jest.fn() },
      } as unknown as React.FormEvent<HTMLInputElement>;
      act(() => {
        (
          wrapper.find("input#disk_text").prop("onChange") as (
            e: React.FormEvent<HTMLInputElement>,
          ) => void
        )(event);
      });
      wrapper.update();
      jest.runAllTimers();

      const sliderChanged = wrapper.find("input").filterWhere(i => i.prop("type") === "range");
      expect(sliderChanged.prop("value")).toBe(2000);
      expect(sliderChanged.prop("max")).toBe(undefined);

      const inputChanged = wrapper.find("input").filterWhere(i => i.prop("type") === "number");
      expect(inputChanged.prop("value")).toBe(2000);
      expect(inputChanged.prop("max")).toBe(undefined);
    });
  });
});
