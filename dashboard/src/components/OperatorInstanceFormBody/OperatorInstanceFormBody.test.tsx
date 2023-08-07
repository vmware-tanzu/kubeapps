// Copyright 2020-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import { act } from "@testing-library/react";
import AlertGroup from "components/AlertGroup";
import ConfirmDialog from "components/ConfirmDialog/ConfirmDialog";
import LoadingWrapper from "components/LoadingWrapper";
import { defaultStore, mountWrapper } from "shared/specs/mountWrapper";
import OperatorAdvancedDeploymentForm from "./OperatorAdvancedDeploymentForm/OperatorAdvancedDeploymentForm";
import OperatorInstanceFormBody, { IOperatorInstanceFormProps } from "./OperatorInstanceFormBody";

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
      fillRect: jest.fn(),
    })),
  });
});

afterEach(() => {
  jest.restoreAllMocks();
});

const defaultProps: IOperatorInstanceFormProps = {
  isFetching: false,
  handleDeploy: jest.fn(),
  defaultValues: "",
};

it("set a loading wrapper", () => {
  const wrapper = mountWrapper(
    defaultStore,
    <OperatorInstanceFormBody {...defaultProps} isFetching={true} />,
  );
  expect(wrapper.find(LoadingWrapper).prop("loaded")).toBe(false);
});

it("set default values", () => {
  const wrapper = mountWrapper(
    defaultStore,
    <OperatorInstanceFormBody {...defaultProps} defaultValues="foo" />,
  );
  expect(wrapper.find(OperatorAdvancedDeploymentForm).prop("appValues")).toBe("foo");
});

it("restores the default values", async () => {
  const wrapper = mountWrapper(
    defaultStore,
    <OperatorInstanceFormBody {...defaultProps} defaultValues="foo" />,
  );

  act(() => {
    (wrapper.find(OperatorAdvancedDeploymentForm).prop("handleValuesChange") as any)("not-foo");
  });
  wrapper.update();
  expect(wrapper.find(OperatorAdvancedDeploymentForm).prop("appValues")).toBe("not-foo");

  const restoreButton = wrapper
    .find(CdsButton)
    .filterWhere(b => b.text().includes("Restore Defaults"));
  act(() => {
    restoreButton.simulate("click");
  });
  act(() => {
    (wrapper.find(ConfirmDialog).prop("onConfirm") as any)();
  });
  wrapper.update();

  expect(wrapper.find(OperatorAdvancedDeploymentForm).prop("appValues")).toBe("foo");
});

it("should submit the form", () => {
  const handleDeploy = jest.fn();
  const wrapper = mountWrapper(
    defaultStore,
    <OperatorInstanceFormBody {...defaultProps} handleDeploy={handleDeploy} />,
  );

  const values = "apiVersion: v1\nmetadata:\n  name: foo";
  act(() => {
    (wrapper.find(OperatorAdvancedDeploymentForm).prop("handleValuesChange") as any)(values);
  });
  const form = wrapper.find("form");
  form.simulate("submit", { preventDefault: jest.fn() });

  const resource = {
    apiVersion: "v1",
    metadata: {
      name: "foo",
    },
  };
  expect(handleDeploy).toHaveBeenCalledWith(resource);
});

it("should catch a syntax error in the form", () => {
  const handleDeploy = jest.fn();
  const wrapper = mountWrapper(
    defaultStore,
    <OperatorInstanceFormBody {...defaultProps} handleDeploy={handleDeploy} />,
  );

  const values = "metadata: invalid!\n  name: foo";
  act(() => {
    (wrapper.find(OperatorAdvancedDeploymentForm).prop("handleValuesChange") as any)(values);
  });
  const form = wrapper.find("form");
  form.simulate("submit", { preventDefault: jest.fn() });

  expect(wrapper.find(AlertGroup)).toIncludeText(
    "Unable parse the resource. Make sure it contains a valid apiVersion",
  );
  expect(handleDeploy).not.toHaveBeenCalled();
});

it("should throw an error if the element doesn't contain an apiVersion", () => {
  const handleDeploy = jest.fn();
  const wrapper = mountWrapper(
    defaultStore,
    <OperatorInstanceFormBody {...defaultProps} handleDeploy={handleDeploy} />,
  );

  const values = "metadata:\nname: foo";
  act(() => {
    (wrapper.find(OperatorAdvancedDeploymentForm).prop("handleValuesChange") as any)(values);
  });
  const form = wrapper.find("form");
  form.simulate("submit", { preventDefault: jest.fn() });

  expect(wrapper.find(AlertGroup)).toIncludeText(
    "Unable parse the resource. Make sure it contains a valid apiVersion",
  );
  expect(handleDeploy).not.toHaveBeenCalled();
});
