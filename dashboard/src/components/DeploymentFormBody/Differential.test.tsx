// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { MonacoDiffEditor } from "react-monaco-editor";
import { SupportedThemes } from "shared/Config";
import { defaultStore, getStore, mountWrapper } from "shared/specs/mountWrapper";
import { IStoreState } from "shared/types";
import Differential from "./Differential";

beforeEach(() => {
  // mock the window.matchMedia for selecting the theme
  Object.defineProperty(window, "matchMedia", {
    writable: true,
    configurable: true,
    value: jest.fn().mockImplementation(query => ({
      matches: false,
      media: query,
      onchange: null,
      addListener: jest.fn(),
      removeListener: jest.fn(),
      addEventListener: jest.fn(),
      removeEventListener: jest.fn(),
      dispatchEvent: jest.fn(),
    })),
  });

  // mock the window.ResizeObserver, required by the MonacoEditor for the layout
  Object.defineProperty(window, "ResizeObserver", {
    writable: true,
    configurable: true,
    value: jest.fn().mockImplementation(() => ({
      observe: jest.fn(),
      unobserve: jest.fn(),
      disconnect: jest.fn(),
    })),
  });

  // mock the window.HTMLCanvasElement.getContext(), required by the MonacoEditor for the layout
  Object.defineProperty(HTMLCanvasElement.prototype, "getContext", {
    writable: true,
    configurable: true,
    value: jest.fn().mockImplementation(() => ({
      clearRect: jest.fn(),
    })),
  });
});

afterEach(() => {
  jest.restoreAllMocks();
});

it("should render a diff between two strings", () => {
  const wrapper = mountWrapper(
    defaultStore,
    <Differential oldValues="foo" newValues="bar" emptyDiffElement={<span>empty</span>} />,
  );
  expect(wrapper.find(MonacoDiffEditor).prop("value")).toBe("bar");
  expect(wrapper.find(MonacoDiffEditor).prop("original")).toBe("foo");
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
  expect(wrapper.find(MonacoDiffEditor).prop("theme")).toBe("light");
});

it("changes theme", () => {
  const wrapper = mountWrapper(
    getStore({ config: { theme: SupportedThemes.dark } } as Partial<IStoreState>),
    <Differential oldValues="foo" newValues="bar" emptyDiffElement={<span>empty</span>} />,
  );
  expect(wrapper.find(MonacoDiffEditor).prop("theme")).toBe("vs-dark");
});
