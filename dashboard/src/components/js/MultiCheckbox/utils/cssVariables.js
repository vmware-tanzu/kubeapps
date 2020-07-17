import { nth } from "lodash-es";

// Css variable for responsive breaks
const cssVarBreakpoint = ["--col-sm-num", "--col-md-num", "--col-lg-num", "--col-xl-num"];

export const assignCssVariables = span => {
  if (span == null) {
    return {
      "--col-sm-num": 1,
      "--col-md-num": 1,
      "--col-lg-num": 1,
      "--col-xl-num": 1,
    };
  } else if (typeof span == "number") {
    return {
      "--col-sm-num": span,
      "--col-md-num": span,
      "--col-lg-num": span,
      "--col-xl-num": span,
    };
  } else {
    let lastValue = span[0];
    return cssVarBreakpoint.reduce((state, next, index) => {
      const value = nth(span, index) || lastValue;
      lastValue = value;
      return {
        ...state,
        [next]: value,
      };
    }, {});
  }
};
