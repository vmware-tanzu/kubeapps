// Copyright 2020-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { defaultStore, mountWrapper } from "shared/specs/mountWrapper";
import { Row } from "./Row";

describe(Row, () => {
  it("renders the content correctly", () => {
    const text = "My Text";
    const wrapper = mountWrapper(defaultStore, <Row>{text}</Row>);

    expect(wrapper.find(Row)).toHaveText(text);
  });

  it("includes the expected CSS class", () => {
    const wrapper = mountWrapper(defaultStore, <Row>Test</Row>);
    expect(wrapper.find(Row).childAt(0)).toHaveClassName("clr-row");
  });

  it("add the role when the row is a list", () => {
    const wrapper = mountWrapper(defaultStore, <Row isList={true}>Test</Row>);
    expect(wrapper.find(Row).childAt(0).prop("role")).toBe("list");
  });
});
