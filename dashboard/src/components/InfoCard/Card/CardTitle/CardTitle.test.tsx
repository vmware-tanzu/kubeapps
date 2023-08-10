// Copyright 2020-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { defaultStore, mountWrapper } from "shared/specs/mountWrapper";
import CardTitle from ".";

describe(CardTitle, () => {
  it("renders the content correctly", () => {
    const text = "My Text";
    const wrapper = mountWrapper(defaultStore, <CardTitle level={1}>{text}</CardTitle>);

    expect(wrapper.find(CardTitle).childAt(0)).toHaveText(text);
  });

  it("includes the expected CSS class", () => {
    const wrapper = mountWrapper(defaultStore, <CardTitle level={1}>Test</CardTitle>);
    expect(wrapper.find(CardTitle).childAt(0)).toHaveClassName("card-title");
  });

  describe("Heading levels", () => {
    [1, 2, 3, 4, 5, 6].forEach(level => {
      it(`renders the h${level} tag based on the level prop`, () => {
        const wrapper = mountWrapper(
          defaultStore,
          <CardTitle level={level as any}>Test</CardTitle>,
        );
        expect(wrapper.find(`h${level}`)).toExist();
      });
    });
  });
});
