// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CustomComponent } from "RemoteComponent";
import { getStore, mountWrapper } from "shared/specs/mountWrapper";
import { IBasicFormParam, IStoreState } from "shared/types";
import CustomFormComponentLoader, { ICustomParamProps } from "./CustomFormParam";

const param: IBasicFormParam = {
  type: "boolean",
  customComponent: {
    className: "test",
  },
  currentValue: true,
  defaultValue: true,
  deployedValue: true,
  hasProperties: false,
  isRequired: false,
  key: "enableMetrics",
  schema: {
    type: "boolean",
  },
  title: "Enable Metrics",
};

const defaultProps = {
  param,
  handleBasicFormParamChange: jest.fn(),
} as ICustomParamProps;

const defaultState = {
  config: { remoteComponentsUrl: "" },
} as IStoreState;

// Ensure remote-component doesn't trigger external requests during this test.
const mockOpen = jest.fn();
const xhrMock: Partial<XMLHttpRequest> = {
  open: mockOpen,
  send: jest.fn(),
  setRequestHeader: jest.fn(),
  readyState: 4,
  status: 200,
  response: "Hello World!",
};

beforeAll((): void => {
  jest.spyOn(window, "XMLHttpRequest").mockImplementation(() => xhrMock as XMLHttpRequest);
});
afterEach((): void => {
  mockOpen.mockReset();
});

it("should render a custom form component", () => {
  const wrapper = mountWrapper(
    getStore(defaultState),
    <CustomFormComponentLoader {...defaultProps} />,
  );
  expect(wrapper.find(CustomFormComponentLoader)).toExist();
});

it("should render the remote component", () => {
  const wrapper = mountWrapper(
    getStore(defaultState),
    <CustomFormComponentLoader {...defaultProps} />,
  );
  expect(wrapper.find(CustomComponent)).toExist();
});

it("should render the remote component with the default URL", () => {
  const wrapper = mountWrapper(
    getStore(defaultState),
    <CustomFormComponentLoader {...defaultProps} />,
  );
  expect(wrapper.find(CustomComponent)).toExist();
  expect(wrapper.find(CustomComponent).prop("url")).toContain("custom_components.js");
});

it("should render the remote component with the URL if set in the config", () => {
  const wrapper = mountWrapper(
    getStore({
      config: { remoteComponentsUrl: "www.thiswebsite.com" },
    } as Partial<IStoreState>),
    <CustomFormComponentLoader {...defaultProps} />,
  );
  expect(wrapper.find(CustomComponent).prop("url")).toBe("www.thiswebsite.com");
  expect(xhrMock.open).toHaveBeenCalledWith("GET", "www.thiswebsite.com", true);
});
