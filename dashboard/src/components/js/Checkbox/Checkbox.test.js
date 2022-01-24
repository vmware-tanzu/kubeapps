// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { shallow } from "enzyme";
import React from "react";
import Checkbox from ".";

const defaultProps = {
  id: "test",
  name: "test",
  label: "I agree",
  value: false,
};

const checkedProps = {
  ...defaultProps,
  value: true,
};

describe(Checkbox, () => {
  it("renders the checkbox and the wrapper", () => {
    const wrapper = shallow(<Checkbox {...defaultProps} />);

    expect(wrapper.find(".clr-checkbox-wrapper")).toExist();
    expect(wrapper.find("input[type='checkbox']")).toExist();
    expect(wrapper.find(".clr-control-label")).toExist();
    expect(wrapper.find(".clr-control-label")).toHaveText(defaultProps.label);
  });

  it("shows the checkbox as checked when the value is true", () => {
    const wrapper = shallow(<Checkbox {...checkedProps} />);

    expect(wrapper.find("input[type='checkbox']").prop("checked")).toBeTruthy();
  });

  it("shows the checkbox as unchecked when the value is false", () => {
    const wrapper = shallow(<Checkbox {...defaultProps} />);

    expect(wrapper.find("input[type='checkbox']").prop("checked")).not.toBeTruthy();
  });
});
