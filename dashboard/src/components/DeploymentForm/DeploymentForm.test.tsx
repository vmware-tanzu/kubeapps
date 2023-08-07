// Copyright 2018-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { CdsSelect } from "@cds/react/select";
import { act } from "@testing-library/react";
import actions from "actions";
import { JSONSchemaType } from "ajv";
import AlertGroup from "components/AlertGroup";
import PackageHeader from "components/PackageHeader/PackageHeader";
import {
  AvailablePackageDetail,
  AvailablePackageReference,
  PackageAppVersion,
  ReconciliationOptions,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages_pb";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins_pb";
import { GetServiceAccountNamesResponse } from "gen/kubeappsapis/plugins/resources/v1alpha1/resources_pb";
import * as ReactRedux from "react-redux";
import * as ReactRouter from "react-router";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { Kube } from "shared/Kube";
import { getStore, initialState, mountWrapper } from "shared/specs/mountWrapper";
import { FetchError, IStoreState, PluginNames } from "shared/types";
import DeploymentForm from "./DeploymentForm";
import DeploymentFormBody from "./DeploymentFormBody";

const defaultProps = {
  pkgName: "foo",
  cluster: "default",
  namespace: "default",
  packageCluster: "default",
  packageNamespace: "kubeapps",
  releaseName: "my-release",
  version: "0.0.1",
  plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
};

const defaultSelectedPkg = {
  versions: [{ appVersion: "10.0.0", pkgVersion: "1.2.3" } as PackageAppVersion],
  availablePackageDetail: {
    name: "test",
    availablePackageRef: {
      identifier: "test/test",
      plugin: { name: "my.plugin", version: "0.0.1" },
    } as AvailablePackageReference,
    defaultValues: "package: defaults",
  } as AvailablePackageDetail,
  pkgVersion: "1.2.4",
  values: "package: defaults",
};

const routePathParam = `/c/${defaultProps.cluster}/ns/${defaultProps.namespace}/apps/new/${defaultProps.plugin.name}/${defaultProps.plugin.version}/${defaultProps.packageCluster}/${defaultProps.packageNamespace}/${defaultProps.pkgName}/versions/${defaultProps.version}`;
const routePath =
  "/c/:cluster/ns/:namespace/apps/new/:pluginName/:pluginVersion/:packageCluster/:packageNamespace/:packageId/versions/:packageVersion";

let spyOnUseDispatch: jest.SpyInstance;
let spyOnUseNavigate: jest.SpyInstance;
let mockNavigate: jest.Func;

beforeEach(() => {
  const mockDispatch = jest.fn().mockReturnValue(true);
  spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
  mockNavigate = jest.fn();
  spyOnUseNavigate = jest.spyOn(ReactRouter, "useNavigate").mockReturnValue(mockNavigate);
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
  spyOnUseDispatch.mockRestore();
  spyOnUseNavigate.mockRestore();
});

it("fetches the available versions", () => {
  const fetchAvailablePackageVersions = jest.fn();
  actions.availablepackages.fetchAndSelectAvailablePackageDetail = fetchAvailablePackageVersions;

  mountWrapper(
    getStore({} as Partial<IStoreState>),
    <MemoryRouter initialEntries={[routePathParam]}>
      <Routes>
        <Route path={routePath} element={<DeploymentForm />} />
      </Routes>
    </MemoryRouter>,
    false,
  );

  expect(fetchAvailablePackageVersions).toHaveBeenCalledWith(
    {
      context: { cluster: defaultProps.packageCluster, namespace: defaultProps.packageNamespace },
      identifier: defaultProps.pkgName,
      plugin: defaultProps.plugin,
    } as AvailablePackageReference,
    defaultProps.version,
  );
});

describe("default values", () => {
  it("uses the selected package default values", () => {
    const wrapper = mountWrapper(
      getStore({ packages: { selected: defaultSelectedPkg } } as IStoreState),
      <MemoryRouter initialEntries={[routePathParam]}>
        <Routes>
          <Route path={routePath} element={<DeploymentForm />} />
        </Routes>
      </MemoryRouter>,
      false,
    );

    expect(wrapper.find(DeploymentFormBody).prop("appValues")).toBe("package: defaults");
  });

  it("displays the default value selections if the package has multiple", () => {
    const state = {
      ...initialState,
      packages: {
        selected: {
          ...defaultSelectedPkg,
          availablePackageDetail: {
            ...defaultSelectedPkg.availablePackageDetail,
            additionalDefaultValues: {
              "values-custom": "custom: values",
              "values-other": "other: values",
            },
          },
        },
      },
    };

    const wrapper = mountWrapper(
      getStore(state),
      <MemoryRouter initialEntries={[routePathParam]}>
        <Routes>
          <Route path={routePath} element={<DeploymentForm />} />
        </Routes>
      </MemoryRouter>,
      false,
    );

    const saSelect = wrapper
      .find(CdsSelect)
      .findWhere(a => a.prop("id") === "defaultValues-selector");

    expect(saSelect).toExist();
    expect(saSelect.find("option").at(0)).not.toHaveProperty("value");
    expect(saSelect.find("option").at(1)).toHaveProp("value", "values-custom");
    expect(saSelect.find("option").at(2)).toHaveProp("value", "values-other");
  });

  it("does not display the default value selections if the package has a single custom values", () => {
    const state = {
      ...initialState,
      packages: {
        selected: {
          ...defaultSelectedPkg,
          availablePackageDetail: {
            ...defaultSelectedPkg.availablePackageDetail,
            additionalDefaultValues: {
              "values-custom": "custom: values",
            },
          },
        },
      },
    };

    const wrapper = mountWrapper(
      getStore(state),
      <MemoryRouter initialEntries={[routePathParam]}>
        <Routes>
          <Route path={routePath} element={<DeploymentForm />} />
        </Routes>
      </MemoryRouter>,
      false,
    );

    const saSelect = wrapper
      .find(CdsSelect)
      .findWhere(a => a.prop("id") === "defaultValues-selector");

    expect(saSelect).not.toExist();
  });
});

describe("renders an error", () => {
  it("renders a custom error if the deployment failed", () => {
    const wrapper = mountWrapper(
      getStore({
        packages: {
          selected: { ...defaultSelectedPkg },
        },
        apps: { error: new Error("wrong format!") },
      } as Partial<IStoreState>),
      <MemoryRouter initialEntries={[routePathParam]}>
        <Routes>
          <Route path={routePath} element={<DeploymentForm />} />
        </Routes>
      </MemoryRouter>,
      false,
    );
    expect(wrapper.find(AlertGroup)).toExist();
    expect(
      wrapper
        .find(AlertGroup)
        .findWhere(a => a.html().includes("An error occurred: wrong format!")),
    ).toExist();
    expect(wrapper.find(PackageHeader)).toExist();
  });

  it("renders a fetch error only", () => {
    const wrapper = mountWrapper(
      getStore({
        packages: { selected: { ...defaultSelectedPkg, error: new FetchError("not found") } },
        apps: { error: undefined },
      } as Partial<IStoreState>),
      <MemoryRouter initialEntries={[routePathParam]}>
        <Routes>
          <Route path={routePath} element={<DeploymentForm />} />
        </Routes>
      </MemoryRouter>,
      false,
    );
    expect(wrapper.find(AlertGroup)).toExist();
    expect(
      wrapper
        .find(AlertGroup)
        .findWhere(a => a.html().includes("Unable to retrieve the package: not found")),
    ).toExist();
    expect(wrapper.find(PackageHeader)).not.toExist();
  });

  it("forwards the appValues when modified", () => {
    const wrapper = mountWrapper(
      getStore({ packages: { selected: defaultSelectedPkg } } as IStoreState),
      <MemoryRouter initialEntries={[routePathParam]}>
        <Routes>
          <Route path={routePath} element={<DeploymentForm />} />
        </Routes>
      </MemoryRouter>,
      false,
    );

    const handleValuesChange: (v: string) => void = wrapper
      .find(DeploymentFormBody)
      .prop("setValues");
    act(() => {
      handleValuesChange("changed: defaults");
    });
    wrapper.update();

    expect(wrapper.find(DeploymentFormBody).prop("appValues")).toBe("changed: defaults");
  });

  it("changes values if the version changes and it has not been modified", () => {
    const wrapper = mountWrapper(
      getStore({ packages: { selected: defaultSelectedPkg } } as IStoreState),
      <MemoryRouter initialEntries={[routePathParam]}>
        <Routes>
          <Route path={routePath} element={<DeploymentForm />} />
        </Routes>
      </MemoryRouter>,
      false,
    );
    expect(wrapper.find(DeploymentFormBody).prop("appValues")).toBe("package: defaults");
  });

  it("display the service account selector", () => {
    const initialEntries = [
      `/c/${defaultProps.cluster}/ns/${defaultProps.namespace}/apps/new/${PluginNames.PACKAGES_KAPP}/${defaultProps.plugin.version}/${defaultProps.packageCluster}/${defaultProps.packageNamespace}/${defaultProps.pkgName}/versions/${defaultProps.version}`,
    ];
    Kube.getServiceAccountNames = jest.fn().mockReturnValue({
      then: jest.fn((f: any) =>
        f({ serviceaccountNames: ["my-sa-1", "my-sa-2"] } as GetServiceAccountNamesResponse),
      ),
      catch: jest.fn(f => f()),
    });

    const wrapper = mountWrapper(
      getStore({ packages: { selected: defaultSelectedPkg } } as IStoreState),
      <MemoryRouter initialEntries={initialEntries}>
        <Routes>
          <Route path={routePath} element={<DeploymentForm />} />
        </Routes>
      </MemoryRouter>,
      false,
    );
    const saSelect = wrapper
      .find(CdsSelect)
      .findWhere(a => a.prop("id") === "serviceaccount-selector");

    expect(saSelect).toExist();
    expect(saSelect.find("option").at(0)).not.toHaveProperty("value");
    expect(saSelect.find("option").at(1)).toHaveProp("value", "my-sa-1");
    expect(saSelect.find("option").at(2)).toHaveProp("value", "my-sa-2");
  });

  it("keep values if the version changes", () => {
    const wrapper = mountWrapper(
      getStore({ packages: { selected: defaultSelectedPkg } } as IStoreState),
      <MemoryRouter initialEntries={[routePathParam]}>
        <Routes>
          <Route path={routePath} element={<DeploymentForm />} />
        </Routes>
      </MemoryRouter>,
      false,
    );

    const handleValuesChange: (v: string) => void = wrapper
      .find(DeploymentFormBody)
      .prop("setValues");
    const setValuesModified: () => void = wrapper
      .find(DeploymentFormBody)
      .prop("setValuesModified");
    act(() => {
      handleValuesChange("foo: bar");
      setValuesModified();
    });
    wrapper.update();
    expect(wrapper.find(DeploymentFormBody).prop("appValues")).toBe("foo: bar");

    wrapper.find("select").simulate("change", { target: { value: "1.2.4" } });
    wrapper.setProps({ selected: { ...defaultSelectedPkg, values: "bar: foo" } });
    wrapper.update();
    expect(wrapper.find(DeploymentFormBody).prop("appValues")).toBe("foo: bar");
  });

  it("triggers a deployment when submitting the form", async () => {
    const installPackage = jest.fn().mockReturnValue(true);
    const navigate = jest.fn();
    actions.installedpackages.installPackage = installPackage;
    spyOnUseNavigate = jest.spyOn(ReactRouter, "useNavigate").mockReturnValue(navigate);

    const appValues = "foo: bar";
    const newAppValues = "foo: modified";
    const schema = { properties: { foo: { type: "string" } } } as unknown as JSONSchemaType<any>;
    const selected = { ...defaultSelectedPkg, values: appValues, schema: schema };

    const wrapper = mountWrapper(
      getStore({ packages: { selected: selected } } as IStoreState),
      <MemoryRouter initialEntries={[routePathParam]}>
        <Routes>
          <Route path={routePath} element={<DeploymentForm />} />
        </Routes>
      </MemoryRouter>,
      false,
    );

    const handleValuesChange: (v: string) => void = wrapper
      .find(DeploymentFormBody)
      .prop("setValues");

    act(() => {
      handleValuesChange(newAppValues);
    });

    wrapper
      .find("#releaseName")
      .simulate("change", { target: { value: defaultProps.releaseName } });

    wrapper.update();

    expect(wrapper.find(DeploymentFormBody).prop("appValues")).toBe(newAppValues);
    expect(wrapper.find(DeploymentForm).find("#releaseName").prop("value")).toBe(
      defaultProps.releaseName,
    );

    await act(async () => {
      // Simulating "submit" causes a console.warning
      await (wrapper.find("form").prop("onSubmit") as (e: any) => Promise<void>)({
        preventDefault: jest.fn(),
      });
    });

    expect(installPackage).toHaveBeenCalledWith(
      defaultProps.cluster,
      defaultProps.namespace,
      defaultSelectedPkg.availablePackageDetail,
      defaultProps.releaseName,
      newAppValues,
      schema,
      {} as ReconciliationOptions,
    );

    expect(navigate).toHaveBeenCalledWith(
      `/c/${defaultProps.cluster}/ns/${defaultProps.namespace}/apps/my.plugin/0.0.1/${defaultProps.releaseName}`,
    );
  });
});
