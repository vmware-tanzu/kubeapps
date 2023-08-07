// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import MonacoEditor from "react-monaco-editor";
import { SupportedThemes } from "shared/Config";
import { defaultStore, getStore, initialState, mountWrapper } from "shared/specs/mountWrapper";
import { IStoreState } from "shared/types";
import AppValues from "./AppValues";

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
      fillRect: jest.fn(),
    })),
  });
});

afterEach(() => {
  jest.restoreAllMocks();
});

it("includes values", () => {
  const wrapper = mountWrapper(defaultStore, <AppValues values="foo: bar" />);
  expect(wrapper.find(MonacoEditor).prop("value")).toBe("foo: bar");
});

it("adds a default message if no values are given", () => {
  const wrapper = mountWrapper(defaultStore, <AppValues values="" />);
  expect(wrapper.find(MonacoEditor)).not.toExist();
  expect(wrapper).toIncludeText(
    "The current application was installed without specifying any values",
  );
});

it("sets light theme by default", () => {
  const wrapper = mountWrapper(defaultStore, <AppValues values="foo: bar" />);
  expect(wrapper.find(MonacoEditor).prop("theme")).toBe("light");
});

it("changes theme", () => {
  const wrapper = mountWrapper(
    getStore({ ...initialState, config: { theme: SupportedThemes.dark } } as Partial<IStoreState>),
    <AppValues values="foo: bar" />,
  );
  expect(wrapper.find(MonacoEditor).prop("theme")).toBe("vs-dark");
});
