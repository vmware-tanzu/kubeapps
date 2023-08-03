// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { shallow } from "enzyme";
import React from "react";
import CardTitle from ".";

describe(CardTitle, () => {
  it("renders the content correctly", () => {
    const text = "My Text";
    const wrapper = shallow(<CardTitle level={1}>{text}</CardTitle>);

    expect(wrapper).toHaveText(text);
    expect(wrapper).toMatchSnapshot();
  });

  it("includes the expected CSS class", () => {
    const wrapper = shallow(<CardTitle level={1}>Test</CardTitle>);
    expect(wrapper).toHaveClassName("card-title");
  });

  describe("Heading levels", () => {
    [1, 2, 3, 4, 5, 6].forEach(level => {
      it(`renders the h${level} tag based on the level prop`, () => {
        const wrapper = shallow(<CardTitle level={level}>Test</CardTitle>);
        expect(wrapper).toHaveDisplayName(`h${level}`);
      });
    });
  });
});
