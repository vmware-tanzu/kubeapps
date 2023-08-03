// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { shallow } from "enzyme";
import React from "react";
import CardBlock from ".";

describe(CardBlock, () => {
  it("renders the content correctly", () => {
    const text = "My Text";
    const wrapper = shallow(<CardBlock>{text}</CardBlock>);

    expect(wrapper).toHaveText(text);
    expect(wrapper).toMatchSnapshot();
  });

  it("includes the expected CSS class", () => {
    const wrapper = shallow(<CardBlock>Test</CardBlock>);
    expect(wrapper).toHaveClassName("card-block");
  });
});
