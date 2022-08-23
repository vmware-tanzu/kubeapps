// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import ReactDiffViewer from "react-diff-viewer";
import { SupportedThemes } from "shared/Config";
import { defaultStore, getStore, mountWrapper } from "shared/specs/mountWrapper";
import { IStoreState } from "shared/types";
import Differential from "./Differential";

it("should render a diff between two strings", () => {
  const wrapper = mountWrapper(
    defaultStore,
    <Differential oldValues="foo" newValues="bar" emptyDiffElement={<span>empty</span>} />,
  );
  expect(wrapper.find(ReactDiffViewer).props()).toMatchObject({ oldValue: "foo", newValue: "bar" });
});

it("should print the emptyDiffText if there are no changes", () => {
  const wrapper = mountWrapper(
    defaultStore,
    <Differential
      oldValues="foo"
      newValues="foo"
      emptyDiffElement={<span>No differences!</span>}
    />,
  );
  expect(wrapper.text()).toMatch("No differences!");
  expect(wrapper.text()).not.toMatch("foo");
});

it("sets light theme by default", () => {
  const wrapper = mountWrapper(
    defaultStore,
    <Differential oldValues="foo" newValues="bar" emptyDiffElement={<span>empty</span>} />,
  );
  expect(wrapper.find(ReactDiffViewer).prop("useDarkTheme")).toBe(false);
});

it("changes theme", () => {
  const wrapper = mountWrapper(
    getStore({ config: { theme: SupportedThemes.dark } } as Partial<IStoreState>),
    <Differential oldValues="foo" newValues="bar" emptyDiffElement={<span>empty</span>} />,
  );
  expect(wrapper.find(ReactDiffViewer).prop("useDarkTheme")).toBe(true);
});
