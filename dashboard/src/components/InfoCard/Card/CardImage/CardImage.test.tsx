// Copyright 2020-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { defaultStore, mountWrapper } from "shared/specs/mountWrapper";
import CardImage from ".";

describe(CardImage, () => {
  it("passes the src and alt attributes to the img", () => {
    const src = "https://example.com";
    const alt = "The Example page";
    const wrapper = mountWrapper(defaultStore, <CardImage src={src} alt={alt} />);

    expect(wrapper.find("img").prop("src")).toBe(src);
    expect(wrapper.find("img").prop("alt")).toBe(alt);
  });

  it("includes the alt property if it's an empty string", () => {
    const view = mountWrapper(defaultStore, <CardImage src="https://example.com" alt="" />);
    expect(view.find("img").prop("alt")).toBe("");
  });

  it("includes the expected CSS class", () => {
    const wrapper = mountWrapper(defaultStore, <CardImage src="https://example.com" alt="" />);
    expect(wrapper.find(CardImage).childAt(0)).toHaveClassName("card-img");
  });
});
