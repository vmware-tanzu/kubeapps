// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { shallow } from "enzyme";
import React from "react";
import CardHeader from ".";

describe(CardHeader, () => {
  it("renders the content correctly", () => {
    const text = "My Text";
    const wrapper = shallow(<CardHeader>{text}</CardHeader>);

    expect(wrapper).toHaveText(text);
    expect(wrapper).toMatchSnapshot();
  });

  it("includes the expected CSS class", () => {
    const wrapper = shallow(<CardHeader>Test</CardHeader>);
    expect(wrapper).toHaveClassName("card-header");
  });

  it("adds the no-border class based on props", () => {
    const wrapper = shallow(<CardHeader noBorder>Test</CardHeader>);
    expect(wrapper).toHaveClassName("no-border");
  });
});
