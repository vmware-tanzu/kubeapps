// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { shallow } from "enzyme";
import { act } from "react-dom/test-utils";
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
    expect(
      wrapper
        .find("input")
        .filterWhere(i => i.prop("type") === "number")
        .prop("value"),
    ).toBe(10);
    expect(
      wrapper
        .find("input")
        .filterWhere(i => i.prop("type") === "range")
        .prop("value"),
    ).toBe(10);
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
    expect(slider.prop("min")).toBe(5);
    expect(slider.prop("max")).toBe(50);
  });
});

it("sets the param minimum to current value if less than min", () => {
  params.forEach(param => {
    const wrapper = shallow(
      <SliderParam
        {...defaultProps}
        param={{ ...param, currentValue: 1, minimum: 100, maximum: 100 }}
      />,
    );

    const slider = wrapper.find("input").filterWhere(i => i.prop("type") === "range");
    expect(slider.prop("min")).toBe(1);
    expect(slider.prop("max")).toBe(100);
  });
});

it("sets the param maximum to current value if greater than", () => {
  params.forEach(param => {
    const wrapper = shallow(
      <SliderParam
        {...defaultProps}
        param={{ ...param, currentValue: 2000, minimum: 100, maximum: 100 }}
      />,
    );

    const slider = wrapper.find("input").filterWhere(i => i.prop("type") === "range");
    expect(slider.prop("min")).toBe(100);
    expect(slider.prop("max")).toBe(2000);
  });
});

it("defaults to the min if the value is undefined", () => {
  params.forEach(param => {
    const wrapper = shallow(
      <SliderParam {...defaultProps} param={{ ...param, currentValue: undefined, minimum: 5 }} />,
    );
    expect(
      wrapper
        .find("input")
        .filterWhere(i => i.prop("type") === "range")
        .prop("value"),
    ).toBe(5);
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

      const event = { currentTarget: { value: "20" } } as React.FormEvent<HTMLInputElement>;
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
      expect(valueChange).toHaveBeenCalledWith(event);
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
      expect(
        wrapper
          .find("input")
          .filterWhere(i => i.prop("type") === "range")
          .prop("value"),
      ).toBe(10);

      const event = { currentTarget: { value: "20" } } as React.FormEvent<HTMLInputElement>;
      act(() => {
        (
          wrapper.find("input#disk_text").prop("onChange") as (
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
      expect(valueChange).toHaveBeenCalledWith(event);
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
      expect(
        wrapper
          .find("input")
          .filterWhere(i => i.prop("type") === "range")
          .prop("value"),
      ).toBe(10);

      const event = { currentTarget: { value: "foo20*#@$" } } as React.FormEvent<HTMLInputElement>;
      act(() => {
        (
          wrapper.find("input#disk_text").prop("onChange") as (
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
      ).toBe(NaN);

      expect(
        wrapper
          .find("input")
          .filterWhere(i => i.prop("type") === "range")
          .prop("value"),
      ).toBe(NaN);

      expect(handleBasicFormParamChange).toHaveBeenCalledWith(param);
      expect(valueChange).toHaveBeenCalledWith(event);
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
      expect(
        wrapper
          .find("input")
          .filterWhere(i => i.prop("type") === "range")
          .prop("value"),
      ).toBe(10);

      const event = { currentTarget: { value: "20.5" } } as React.FormEvent<HTMLInputElement>;
      act(() => {
        (
          wrapper.find("input#disk_text").prop("onChange") as (
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
      ).toBe(20.5);

      expect(
        wrapper
          .find("input")
          .filterWhere(i => i.prop("type") === "range")
          .prop("value"),
      ).toBe(20.5);

      expect(handleBasicFormParamChange).toHaveBeenCalledWith(param);
      expect(valueChange).toHaveBeenCalledWith(event);
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
      expect(
        wrapper
          .find("input")
          .filterWhere(i => i.prop("type") === "range")
          .prop("value"),
      ).toBe(10);

      const event = { currentTarget: { value: "2000" } } as React.FormEvent<HTMLInputElement>;
      act(() => {
        (
          wrapper.find("input#disk_text").prop("onChange") as (
            e: React.FormEvent<HTMLInputElement>,
          ) => void
        )(event);
      });
      wrapper.update();
      jest.runAllTimers();

      expect(
        wrapper
          .find("input")
          .filterWhere(i => i.prop("type") === "range")
          .prop("value"),
      ).toBe(2000);
      const slider = wrapper.find("input").filterWhere(i => i.prop("type") === "range");
      expect(slider.prop("max")).toBe(2000);
    });
  });
});
