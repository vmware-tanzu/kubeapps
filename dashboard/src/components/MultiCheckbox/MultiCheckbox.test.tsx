// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsCheckbox } from "@cds/react/checkbox";
import { shallow } from "enzyme";
import { MultiCheckbox } from "./MultiCheckbox";

const defaultProps = {
  name: "test",
  options: ["yes", "no", "maybe"],
  value: [],
  span: 1,
};

describe(MultiCheckbox, () => {
  it("renders the wrappers, labels and inputs", () => {
    const wrapper = shallow(<MultiCheckbox {...defaultProps} />);
    expect(wrapper.find(CdsCheckbox).length).toBe(defaultProps.options.length);
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
});
