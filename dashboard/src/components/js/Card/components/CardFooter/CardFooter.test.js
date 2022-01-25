// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { shallow } from "enzyme";
import React from "react";
import CardFooter from ".";

describe(CardFooter, () => {
  it("renders the content correctly", () => {
    const text = "My Text";
    const wrapper = shallow(<CardFooter>{text}</CardFooter>);

    expect(wrapper).toHaveText(text);
    expect(wrapper).toMatchSnapshot();
  });

  it("includes the expected CSS class", () => {
    const wrapper = shallow(<CardFooter>Test</CardFooter>);
    expect(wrapper).toHaveClassName("card-footer");
  });
});
