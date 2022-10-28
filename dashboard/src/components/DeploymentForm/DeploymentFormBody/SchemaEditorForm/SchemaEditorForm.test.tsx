// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { SupportedThemes } from "shared/Config";
import { defaultStore, getStore, mountWrapper } from "shared/specs/mountWrapper";
import { IStoreState } from "shared/types";
import SchemaEditorForm, { ISchemaEditorForm } from "./SchemaEditorForm";

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

  // mock the window.ResizeObserver, required by the MonacoDiffEditor for the layout
  Object.defineProperty(window, "ResizeObserver", {
    writable: true,
    configurable: true,
    value: jest.fn().mockImplementation(() => ({
      observe: jest.fn(),
      unobserve: jest.fn(),
      disconnect: jest.fn(),
    })),
  });

  // mock the window.HTMLCanvasElement.getContext(), required by the MonacoDiffEditor for the layout
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

const defaultProps: ISchemaEditorForm = {
  handleValuesChange: jest.fn(),
  schemaFromTheAvailablePackage: "{}",
  schemaFromTheParentContainer: "{}",
};

it("includes values", () => {
  const wrapper = mountWrapper(
    defaultStore,
    <SchemaEditorForm {...defaultProps} schemaFromTheParentContainer='{ "type": "string" }' />,
  );
  expect(wrapper.find("MonacoDiffEditor").prop("value")).toBe('{ "type": "string" }');
});

it("sets light theme by default", () => {
  const wrapper = mountWrapper(defaultStore, <SchemaEditorForm {...defaultProps} />);
  expect(wrapper.find("MonacoDiffEditor").prop("theme")).toBe("light");
});

it("changes theme", () => {
  const wrapper = mountWrapper(
    getStore({ config: { theme: SupportedThemes.dark } } as Partial<IStoreState>),
    <SchemaEditorForm {...defaultProps} />,
  );
  expect(wrapper.find("MonacoDiffEditor").prop("theme")).toBe("vs-dark");
});
