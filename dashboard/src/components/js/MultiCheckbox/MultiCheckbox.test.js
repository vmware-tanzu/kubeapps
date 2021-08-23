import { shallow } from "enzyme";
import React from "react";
import MultiCheckbox from ".";

const defaultProps = {
  options: ["yes", "no", "maybe"],
  value: [],
  span: 1,
};

describe(MultiCheckbox, () => {
  it("renders the wrappers, labels and inputs", () => {
    const wrapper = shallow(<MultiCheckbox {...defaultProps} />);

    expect(wrapper.find(".clr-checkbox-wrapper").length).toBe(defaultProps.options.length);
    expect(wrapper.find("input[type='checkbox']").length).toBe(defaultProps.options.length);
  });

  it("shows the status of the checkboxes based on default values", () => {
    const propsWithValue = {
      ...defaultProps,
      value: ["yes", "maybe"],
    };
    const wrapper = shallow(<MultiCheckbox {...propsWithValue} />);

    expect(wrapper.find("input[value='yes']").prop("checked")).toBeTruthy();
    expect(wrapper.find("input[value='no']").prop("checked")).toBeFalsy();
    expect(wrapper.find("input[value='maybe']").prop("checked")).toBeTruthy();
  });

  describe("Multiple Columns", () => {
    it("allows multiple columns", () => {
      const propsWithColumns = {
        ...defaultProps,
        span: 2,
      };

      const wrapper = shallow(<MultiCheckbox {...propsWithColumns} />);

      expect(wrapper.find(".multicheckbox-wrapper").first().prop("style")).toStrictEqual({
        "--col-lg-num": 2,
        "--col-md-num": 2,
        "--col-sm-num": 2,
        "--col-xl-num": 2,
      });
    });

    it("default to 1 column if the property is not specified", () => {
      const propsWithColumns = {
        ...defaultProps,
        span: undefined,
      };

      const wrapper = shallow(<MultiCheckbox {...propsWithColumns} />);

      expect(wrapper.find(".multicheckbox-wrapper").first().prop("style")).toStrictEqual({
        "--col-lg-num": 1,
        "--col-md-num": 1,
        "--col-sm-num": 1,
        "--col-xl-num": 1,
      });
    });

    it("allows multiple columns with different sizes", () => {
      const propsWithColumns = {
        ...defaultProps,
        span: [1, 2, 3, 3],
      };

      const wrapper = shallow(<MultiCheckbox {...propsWithColumns} />);

      expect(wrapper.find(".multicheckbox-wrapper").first().prop("style")).toStrictEqual({
        "--col-lg-num": 3,
        "--col-md-num": 2,
        "--col-sm-num": 1,
        "--col-xl-num": 3,
      });
    });

    it("allows multiple columns with missing sizes", () => {
      const propsWithColumns = {
        ...defaultProps,
        span: [1, 2],
      };
      const wrapper = shallow(<MultiCheckbox {...propsWithColumns} />);

      expect(wrapper.find(".multicheckbox-wrapper").first().prop("style")).toStrictEqual({
        "--col-lg-num": 2,
        "--col-md-num": 2,
        "--col-sm-num": 1,
        "--col-xl-num": 2,
      });
    });
  });
});
