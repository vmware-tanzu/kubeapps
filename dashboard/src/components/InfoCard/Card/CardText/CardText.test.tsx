// Copyright 2020-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { defaultStore, mountWrapper } from "shared/specs/mountWrapper";
import CardText from ".";

describe(CardText, () => {
  it("renders the content correctly", () => {
    const text = "My Text";
    const wrapper = mountWrapper(defaultStore, <CardText>{text}</CardText>);

    expect(wrapper.find(CardText).childAt(0)).toHaveText(text);
  });

  it("includes the expected CSS class", () => {
    const wrapper = mountWrapper(defaultStore, <CardText>Test</CardText>);
    expect(wrapper.find(CardText).childAt(0)).toHaveClassName("card-text");
  });
});
