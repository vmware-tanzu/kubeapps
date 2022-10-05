// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { JSONSchemaType } from "ajv";
import {
  AvailablePackageDetail,
  PackageAppVersion,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { act } from "react-dom/test-utils";
import { MonacoDiffEditor } from "react-monaco-editor";
import { defaultStore, mountWrapper } from "shared/specs/mountWrapper";
import { IPackageState } from "shared/types";
import BasicDeploymentForm from "./BasicDeploymentForm";
import DeploymentFormBody, { IDeploymentFormBodyProps } from "./DeploymentFormBody";

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

const defaultProps: IDeploymentFormBodyProps = {
  deploymentEvent: "install",
  packageId: "foo",
  packageVersion: "1.0.0",
  packagesIsFetching: false,
  selected: {} as IPackageState["selected"],
  appValues: "foo: bar\n",
  setValues: jest.fn(),
  setValuesModified: jest.fn(),
  formRef: { current: null },
};

jest.useFakeTimers();

const versions = [{ appVersion: "10.0.0", pkgVersion: "1.2.3" }] as PackageAppVersion[];

// Note that most of the tests that cover DeploymentFormBody component are in
// in the DeploymentForm and UpgradeForm parent components

// Context at https://github.com/vmware-tanzu/kubeapps/issues/1293
it("should modify the original values of the differential component if parsed as YAML object", () => {
  const oldValues = `a: b


c: d
`;
  const schema = {
    properties: { a: { type: "string", form: true } },
  } as unknown as JSONSchemaType<any>;
  const selected = {
    values: oldValues,
    schema,
    versions: [versions[0], { ...versions[0], pkgVersion: "1.2.4" } as PackageAppVersion],
    availablePackageDetail: { name: "my-version" } as AvailablePackageDetail,
  } as IPackageState["selected"];

  const wrapper = mountWrapper(
    defaultStore,
    <DeploymentFormBody {...defaultProps} selected={selected} />,
  );
  expect(wrapper.find(MonacoDiffEditor).prop("original")).toBe(oldValues);

  // Trigger a change in the basic form and a YAML parse
  const input = wrapper
    .find(BasicDeploymentForm)
    .find("input")
    .filterWhere(i => i.prop("id") === "a"); // the input for the property "a"

  act(() => {
    input.simulate("change", { currentTarget: "e" });
    jest.advanceTimersByTime(500);
  });
  wrapper.update();

  const expectedValues = `a: b


c: d
`;
  expect(wrapper.find(MonacoDiffEditor).prop("original")).toBe(expectedValues);
});
