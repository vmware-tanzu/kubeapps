// Copyright 2020-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { defaultStore, mountWrapper } from "shared/specs/mountWrapper";
import CardBlock from ".";

describe(CardBlock, () => {
  it("renders the content correctly", () => {
    const text = "My Text";
    const wrapper = mountWrapper(defaultStore, <CardBlock>{text}</CardBlock>);

    expect(wrapper.find(CardBlock).childAt(0)).toHaveText(text);
  });

  it("includes the expected CSS class", () => {
    const wrapper = mountWrapper(defaultStore, <CardBlock>Test</CardBlock>);
    expect(wrapper.find(CardBlock).childAt(0)).toHaveClassName("card-block");
  });
});
