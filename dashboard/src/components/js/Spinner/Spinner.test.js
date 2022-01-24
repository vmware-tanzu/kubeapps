// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { shallow } from "enzyme";
import React from "react";
import Spinner from ".";

describe(Spinner, () => {
  it("displays the basic spinner", () => {
    const wrapper = shallow(<Spinner />);
    expect(wrapper).toMatchSnapshot();
  });

  it("sets a different text", () => {
    const text = "My Text";
    const wrapper = shallow(<Spinner text={text} />);

    expect(wrapper).toHaveText(text);
  });

  it("applies the center style based on props", () => {
    const wrapper = shallow(<Spinner center />);

    expect(wrapper).toHaveClassName("spinner-center");
  });

  it("applies the inline style based on props", () => {
    const wrapper = shallow(<Spinner inline />);

    expect(wrapper.children().first()).toHaveClassName("spinner-inline");
  });

  it("ignores the center style if the inline mode is enabled", () => {
    const wrapper = shallow(<Spinner center inline />);

    expect(wrapper).not.toHaveClassName("spinner-center");
  });

  it("applies the inverse style based on props", () => {
    const wrapper = shallow(<Spinner inverse />);

    expect(wrapper.children().first()).toHaveClassName("spinner-inverse");
  });

  it("applies the small style based on props", () => {
    const wrapper = shallow(<Spinner small />);

    expect(wrapper.children().first()).toHaveClassName("spinner-sm");
  });

  it("applies the medium style based on props", () => {
    const wrapper = shallow(<Spinner medium />);

    expect(wrapper.children().first()).toHaveClassName("spinner-md");
  });
});
