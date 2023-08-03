// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { render, shallow } from "enzyme";
import React from "react";
import CardImage from ".";

describe(CardImage, () => {
  it("passes the src and alt attributes to the img", () => {
    const src = "https://example.com";
    const alt = "The Example page";
    const wrapper = shallow(<CardImage src={src} alt={alt} />);

    expect(wrapper.find("img").prop("src")).toBe(src);
    expect(wrapper.find("img").prop("alt")).toBe(alt);
  });

  it("includes the alt property if it's an empty string", () => {
    const view = render(<CardImage src="https://example.com" alt="" />);
    expect(view.find("img").attr("alt")).toBe("");
  });

  it("includes the expected CSS class", () => {
    const wrapper = shallow(<CardImage src="https://example.com" alt="" />);
    expect(wrapper).toHaveClassName("card-img");
  });
});
