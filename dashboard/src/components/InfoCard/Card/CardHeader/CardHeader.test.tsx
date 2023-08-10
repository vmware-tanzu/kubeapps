// Copyright 2020-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { mountWrapper, defaultStore } from "shared/specs/mountWrapper";
import CardHeader from ".";

describe(CardHeader, () => {
  it("renders the content correctly", () => {
    const text = "My Text";
    const wrapper = mountWrapper(defaultStore, <CardHeader>{text}</CardHeader>);
    expect(wrapper.find(CardHeader).childAt(0)).toHaveText(text);
  });

  it("includes the expected CSS class", () => {
    const wrapper = mountWrapper(defaultStore, <CardHeader>Test</CardHeader>);
    expect(wrapper.find(CardHeader).childAt(0)).toHaveClassName("card-header");
  });

  it("adds the no-border class based on props", () => {
    const wrapper = mountWrapper(defaultStore, <CardHeader noBorder>Test</CardHeader>);
    expect(wrapper.find(CardHeader).childAt(0)).toHaveClassName("no-border");
  });
});
