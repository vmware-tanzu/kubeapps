// Copyright 2019-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsButton } from "@cds/react/button";
import { act } from "@testing-library/react";
import { JSONSchemaType } from "ajv";
import {
  AvailablePackageDetail,
  PackageAppVersion,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages_pb";
import { MonacoDiffEditor } from "react-monaco-editor";
import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";
import { defaultStore, getStore, initialState, mountWrapper } from "shared/specs/mountWrapper";
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
      fillRect: jest.fn(),
    })),
  });
});

// Mocking react-monaco-editor to a simple empty <div> to prevent issues with Jest
// otherwise, an error with while registering the diff webworker is thrown
// rel: https://github.com/microsoft/vscode/pull/192151
jest.mock("react-monaco-editor", () => {
  return {
    MonacoDiffEditor: () => <div />,
  };
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
      fillRect: jest.fn(),
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

const defaultSchema = {
  properties: { a: { type: "string" } },
} as unknown as JSONSchemaType<any>;

const defaultValues = `a: b


c: d
`;

const versions = [{ appVersion: "10.0.0", pkgVersion: "1.2.3" }] as PackageAppVersion[];
const selected = {
  values: defaultValues,
  schema: defaultSchema,
  versions: [versions[0], { ...versions[0], pkgVersion: "1.2.4" } as PackageAppVersion],
  availablePackageDetail: { name: "my-version" } as AvailablePackageDetail,
} as IPackageState["selected"];

// Note that most of the tests that cover DeploymentFormBody component are in
// in the DeploymentForm and UpgradeForm parent components

// Context at https://github.com/vmware-tanzu/kubeapps/issues/1293
it("should modify the original values of the differential component if parsed as YAML object", () => {
  const wrapper = mountWrapper(
    defaultStore,
    <DeploymentFormBody {...defaultProps} selected={{ ...selected, values: defaultValues }} />,
  );

  expect(
    wrapper
      .find(MonacoDiffEditor)
      .filterWhere(p => p.prop("language") === "yaml")
      .prop("original"),
  ).toBe(defaultValues);

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
  expect(
    wrapper
      .find(MonacoDiffEditor)
      .filterWhere(p => p.prop("language") === "yaml")
      .prop("original"),
  ).toBe(expectedValues);
});

it("should not render a schema editor if the feature flag is disabled", () => {
  const state = {
    ...initialState,
    config: {
      ...initialState.config,
      featureFlags: { ...initialState.config.featureFlags, schemaEditor: { enabled: false } },
    },
  };
  const wrapper = mountWrapper(
    getStore(state),
    <DeploymentFormBody {...defaultProps} selected={{ ...selected, schema: defaultSchema }} />,
  );

  expect(
    wrapper.find(MonacoDiffEditor).filterWhere(p => p.prop("language") === "json"),
  ).not.toExist();
});

// Reproduce https://github.com/vmware-tanzu/kubeapps/issues/5805
it("should not render a schema editor if the feature flag is not set", () => {
  const { schemaEditor, ...featureFlagsWithout } = initialState.config.featureFlags;
  const state = {
    ...initialState,
    config: {
      ...initialState.config,
      featureFlags: featureFlagsWithout,
    },
  };
  const mockStore = configureMockStore([thunk]);
  const wrapper = mountWrapper(
    mockStore(state),
    <DeploymentFormBody {...defaultProps} selected={{ ...selected, schema: defaultSchema }} />,
  );

  expect(
    wrapper.find(MonacoDiffEditor).filterWhere(p => p.prop("language") === "json"),
  ).not.toExist();
});
it("should render a schema editor if the feature flag is enabled", () => {
  const state = {
    ...initialState,
    config: {
      ...initialState.config,
      featureFlags: { ...initialState.config.featureFlags, schemaEditor: { enabled: true } },
    },
  };
  const wrapper = mountWrapper(
    getStore(state),
    <DeploymentFormBody {...defaultProps} selected={{ ...selected, schema: defaultSchema }} />,
  );

  const expectedSchema = `{
  "properties": {
    "a": {
      "type": "string"
    }
  }
}`;

  // find the schema editor
  expect(
    wrapper
      .find(MonacoDiffEditor)
      .filterWhere(p => p.prop("language") === "json")
      .prop("original"),
  ).toBe(expectedSchema);

  // ensure the schema is being rendered as a basic form
  expect(
    wrapper
      .find(BasicDeploymentForm)
      .find("input")
      .filterWhere(i => i.prop("id") === "a"), // the input for the property "a"
  ).toExist();

  // ensure there is no button to update the schema if not modified
  expect(
    wrapper.find(CdsButton).filterWhere(b => b.text().includes("Update schema")),
  ).not.toExist();

  const newSchema = `{
    "properties": {
      "changedPropertyName": {
        "type": "string"
      }
    }
  }`;

  // update the schema
  act(() => {
    (
      wrapper
        .find(MonacoDiffEditor)
        .filterWhere(p => p.prop("language") === "json")
        .prop("onChange") as any
    )(newSchema);
  });
  wrapper.update();

  // ensure the new schema is in the editor
  expect(
    wrapper
      .find(MonacoDiffEditor)
      .filterWhere(p => p.prop("language") === "json")
      .prop("original"),
  ).toBe(newSchema);

  // click on the button to update the basic form
  act(() => {
    wrapper
      .find(CdsButton)
      .filterWhere(b => b.text().includes("Update schema"))
      .simulate("click");
  });
  wrapper.update();

  // ensure the basic form has been updated
  expect(
    wrapper
      .find(BasicDeploymentForm)
      .find("input")
      .filterWhere(i => i.prop("id") === "changedPropertyName"),
  ).toExist();
});
