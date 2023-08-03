// Copyright 2020-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { defaultStore, mountWrapper } from "shared/specs/mountWrapper";
import CardFooter from ".";

describe(CardFooter, () => {
  it("renders the content correctly", () => {
    const text = "My Text";
    const wrapper = mountWrapper(defaultStore, <CardFooter>{text}</CardFooter>);

    expect(wrapper.find(CardFooter).childAt(0)).toHaveText(text);
  });

  it("includes the expected CSS class", () => {
    const wrapper = mountWrapper(defaultStore, <CardFooter>Test</CardFooter>);
    expect(wrapper.find(CardFooter).childAt(0)).toHaveClassName("card-footer");
  });
});
