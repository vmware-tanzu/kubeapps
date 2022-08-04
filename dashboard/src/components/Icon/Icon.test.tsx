// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { act } from "react-dom/test-utils";
import { defaultStore, mountWrapper } from "shared/specs/mountWrapper";
import Icon from "./Icon";

it("should render an icon", () => {
  const wrapper = mountWrapper(defaultStore, <Icon icon="foo" />);
  expect(wrapper.find("img").prop("src")).toBe("foo");
});

it("should use the default icon if not given", () => {
  const wrapper = mountWrapper(defaultStore, <Icon />);
  expect(wrapper.find("img").prop("src")).toBe("placeholder.svg");
});

it("should fallback to the placeholder if an error happens", () => {
  const wrapper = mountWrapper(defaultStore, <Icon icon="foo" />);
  expect(wrapper.find("img").prop("src")).toBe("foo");
  act(() => {
    wrapper.find("img").simulate("error", {});
  });
  wrapper.update();
  expect(wrapper.find("img").prop("src")).toBe("placeholder.svg");
});
