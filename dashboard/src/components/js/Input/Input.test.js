// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { shallow } from "enzyme";
import React from "react";
import Input from ".";

const defaultProps = {
  name: "test",
};

const allProps = {
  ...defaultProps,
  type: "text",
  placeholder: "test",
};

describe(Input, () => {
  it("renders the expected items", () => {
    const wrapper = shallow(<Input {...defaultProps} />);

    expect(wrapper.find(".clr-input-wrapper")).toExist();
    expect(wrapper.find(".clr-input")).toExist();
  });

  it("passes all the props to the input", () => {
    const wrapper = shallow(<Input {...allProps} />);
    const input = wrapper.find("input");

    expect(input.prop("placeholder")).toBe(allProps.placeholder);
    expect(input.prop("type")).toBe(allProps.type);
    expect(input.prop("name")).toBe(allProps.name);
  });

  it("renders a children if it's passed", () => {
    const children = <p>help</p>;
    const wrapper = shallow(<Input {...defaultProps}>{children}</Input>);

    expect(wrapper.find("p")).toExist();
    expect(wrapper.find("p")).toHaveText("help");
  });
});
