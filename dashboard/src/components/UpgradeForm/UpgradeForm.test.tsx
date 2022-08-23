// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import actions from "actions";
import DeploymentFormBody from "components/DeploymentFormBody/DeploymentFormBody";
import Alert from "components/js/Alert";
import LoadingWrapper from "components/LoadingWrapper/LoadingWrapper";
import PackageHeader from "components/PackageHeader/PackageHeader";
import PackageVersionSelector from "components/PackageHeader/PackageVersionSelector";
import {
  AvailablePackageDetail,
  AvailablePackageReference,
  Context,
  InstalledPackageReference,
  InstalledPackageStatus,
  InstalledPackageStatus_StatusReason,
  Maintainer,
  PackageAppVersion,
  VersionReference,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import { cloneDeep } from "lodash";
import { act } from "react-dom/test-utils";
import * as ReactRedux from "react-redux";
import { MemoryRouter, Route } from "react-router-dom";
import PackagesService from "shared/PackagesService";
import { defaultStore, getStore, mountWrapper } from "shared/specs/mountWrapper";
import {
  CustomInstalledPackageDetail,
  FetchError,
  IInstalledPackageState,
  IPackageState,
  IStoreState,
} from "shared/types";
import * as url from "shared/url";
import UpgradeForm from "./UpgradeForm";

const testVersion: PackageAppVersion = {
  pkgVersion: "1.2.3",
  appVersion: "4.5.6",
};

const schema = { properties: { foo: { type: "string" } } };

const defaultProps = {
  packageId: "stable/bar",
  namespace: "default",
  cluster: "default",
  releaseName: "my-release",
  repoNamespace: "kubeapps",
  plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
};

const availablePkgDetail = {
  name: "foo",
  categories: [""],
  displayName: "foo",
  iconUrl: "https://icon.com",
  repoUrl: "https://repo.com",
  homeUrl: "https://example.com",
  sourceUrls: ["test"],
  shortDescription: "test",
  longDescription: "test",
  availablePackageRef: {
    identifier: "foo/foo",
    context: { cluster: "", namespace: "package-namespace" } as Context,
    plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
  },
  valuesSchema: '"$schema": "http://json-schema.org/schema#"',
  defaultValues: "default: values",
  maintainers: [{ name: "test", email: "test" }] as Maintainer[],
  readme: "test",
  version: {
    appVersion: testVersion.appVersion,
    pkgVersion: testVersion.pkgVersion,
  } as PackageAppVersion,
} as AvailablePackageDetail;

const installedPkgDetail = {
  name: "test",
  postInstallationNotes: "test",
  valuesApplied: "foo:",
  availablePackageRef: {
    identifier: "stable/bar",
    context: { cluster: defaultProps.cluster, namespace: defaultProps.repoNamespace } as Context,
    plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
  } as AvailablePackageReference,
  currentVersion: { appVersion: "10.0.0", pkgVersion: "1.0.0" } as PackageAppVersion,
  installedPackageRef: {
    identifier: "stable/bar",
    pkgVersion: "1.0.0",
    context: { cluster: defaultProps.cluster, namespace: defaultProps.repoNamespace } as Context,
    plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
  } as InstalledPackageReference,
  latestMatchingVersion: { appVersion: "10.0.0", pkgVersion: "1.0.0" } as PackageAppVersion,
  latestVersion: { appVersion: "10.0.0", pkgVersion: "1.0.0" } as PackageAppVersion,
  pkgVersionReference: { version: "1" } as VersionReference,
  reconciliationOptions: {},
  status: {
    ready: true,
    reason: InstalledPackageStatus_StatusReason.STATUS_REASON_INSTALLED,
    userReason: "deployed",
  } as InstalledPackageStatus,
} as CustomInstalledPackageDetail;

const selectedPkg = {
  availablePackageDetail: availablePkgDetail,
  pkgVersion: testVersion.pkgVersion,
  appVersion: testVersion.appVersion,
  readme: "readme",
  values: "initial: values",
  versions: [testVersion],
  schema: schema as any,
};

const routePathParam = `/c/${defaultProps.cluster}/ns/${defaultProps.namespace}/apps/${defaultProps.plugin.name}/${defaultProps.plugin.version}/${defaultProps.releaseName}/upgrade`;
const routePath = "/c/:cluster/ns/:namespace/apps/:pluginName/:pluginVersion/:releaseName/upgrade";

describe("it behaves like a loading component", () => {
  it("if the app is being fetched", () => {
    const state = {
      ...defaultStore,
      apps: {
        isFetching: true,
      } as IInstalledPackageState,
    };
    expect(
      mountWrapper(
        getStore({ ...state } as Partial<IStoreState>),
        <MemoryRouter initialEntries={[routePathParam]}>
          <Route path={routePath}>
            <UpgradeForm />,
          </Route>
        </MemoryRouter>,
      ).find(LoadingWrapper),
    ).toExist();
  });

  it("if the package is being fetched", () => {
    const state = {
      ...defaultStore,
      packages: {
        isFetching: true,
      } as IPackageState,
    };
    expect(
      mountWrapper(
        getStore({ ...state } as Partial<IStoreState>),
        <MemoryRouter initialEntries={[routePathParam]}>
          <Route path={routePath}>
            <UpgradeForm />,
          </Route>
        </MemoryRouter>,
      ).find(LoadingWrapper),
    ).toExist();
  });
  it("if there are no versions", () => {
    const state = {
      ...defaultStore,
      packages: {
        selected: {
          versions: [] as PackageAppVersion[],
        },
      } as IPackageState,
    };
    expect(
      mountWrapper(
        getStore({ ...state } as Partial<IStoreState>),
        <MemoryRouter initialEntries={[routePathParam]}>
          <Route path={routePath}>
            <UpgradeForm />,
          </Route>
        </MemoryRouter>,
      ).find(LoadingWrapper),
    ).toExist();
  });

  it("if there is no version", () => {
    const state = {
      ...defaultStore,
      packages: {
        selected: {
          availablePackageDetail: undefined,
        },
      } as IPackageState,
    };

    expect(
      mountWrapper(
        getStore({ ...state } as Partial<IStoreState>),
        <MemoryRouter initialEntries={[routePathParam]}>
          <Route path={routePath}>
            <UpgradeForm />,
          </Route>
        </MemoryRouter>,
      ).find(LoadingWrapper),
    ).toExist();
  });
});

it("fetches the available versions", () => {
  const getAvailablePackageVersions = jest.fn();
  PackagesService.getAvailablePackageVersions = getAvailablePackageVersions;

  const state = {
    ...defaultStore,
    apps: {
      selected: installedPkgDetail,
      selectedDetails: availablePkgDetail,
      isFetching: false,
    } as IInstalledPackageState,
  };
  mountWrapper(
    getStore({ ...state } as Partial<IStoreState>),
    <MemoryRouter initialEntries={[routePathParam]}>
      <Route path={routePath}>
        <UpgradeForm />,
      </Route>
    </MemoryRouter>,
  );
  expect(getAvailablePackageVersions).toHaveBeenCalledWith({
    context: {
      cluster: defaultProps.cluster,
      namespace: defaultProps.repoNamespace,
    },
    identifier: defaultProps.packageId,
    plugin: defaultProps.plugin,
  } as AvailablePackageReference);
});

it("hides the PackageVersionSelector in the PackageHeader", () => {
  const state = {
    ...defaultStore,
    apps: {
      selected: installedPkgDetail,
      selectedDetails: availablePkgDetail,
      isFetching: false,
    } as IInstalledPackageState,
    packages: {
      selected: selectedPkg,
    } as IPackageState,
  };
  const wrapper = mountWrapper(
    getStore({ ...state } as Partial<IStoreState>),
    <MemoryRouter initialEntries={[routePathParam]}>
      <Route path={routePath}>
        <UpgradeForm />,
      </Route>
    </MemoryRouter>,
  );
  expect(wrapper.find(PackageVersionSelector)).toHaveLength(1);
  expect(wrapper.find(PackageHeader)).toHaveProp("hideVersionsSelector", true);
});

it("does not fetch the current package version if there is already one in the state", () => {
  const getAvailablePackageDetail = jest.fn();
  PackagesService.getAvailablePackageDetail = getAvailablePackageDetail;
  const state = {
    ...defaultStore,
    apps: {
      selected: installedPkgDetail,
      selectedDetails: availablePkgDetail,
      isFetching: false,
    } as IInstalledPackageState,
    packages: {
      selected: selectedPkg,
    } as IPackageState,
  };
  mountWrapper(
    getStore({ ...state } as Partial<IStoreState>),
    <MemoryRouter initialEntries={[routePathParam]}>
      <Route path={routePath}>
        <UpgradeForm />,
      </Route>
    </MemoryRouter>,
  );
  expect(getAvailablePackageDetail).not.toHaveBeenCalled();
});

describe("renders an error", () => {
  it("renders an alert if the deployment failed", () => {
    const state = {
      ...defaultStore,
      apps: {
        selected: installedPkgDetail,
        selectedDetails: availablePkgDetail,
        isFetching: false,
        error: new FetchError("wrong format!"),
      } as IInstalledPackageState,
      packages: {
        selected: {
          availablePackageDetail: availablePkgDetail,
          pkgVersion: testVersion.pkgVersion,
          appVersion: testVersion.appVersion,
          readme: "readme",
          values: "initial: values",
          versions: [testVersion],
          schema: schema as any,
        },
      } as IPackageState,
    };

    const wrapper = mountWrapper(
      getStore({ ...state } as Partial<IStoreState>),
      <MemoryRouter initialEntries={[routePathParam]}>
        <Route path={routePath}>
          <UpgradeForm />,
        </Route>
      </MemoryRouter>,
    );
    expect(wrapper.find(Alert).exists()).toBe(true);
    expect(wrapper.find(Alert).first()).toIncludeText("wrong format!");
  });
});

it("empty values applied is allowed", () => {
  const installedPackageDetails = cloneDeep(installedPkgDetail);
  installedPackageDetails.valuesApplied = "";
  const state = {
    ...defaultStore,
    apps: {
      selected: installedPackageDetails,
      selectedDetails: availablePkgDetail,
      isFetching: false,
    } as IInstalledPackageState,
    packages: {
      selected: selectedPkg,
    } as IPackageState,
  };

  const wrapper = mountWrapper(
    getStore({ ...state } as Partial<IStoreState>),
    <MemoryRouter initialEntries={[routePathParam]}>
      <Route path={routePath}>
        <UpgradeForm />,
      </Route>
    </MemoryRouter>,
  );
  expect(wrapper.find(DeploymentFormBody).prop("packageVersion")).toBe("1.0.0");
});

it("defaults the upgrade version to the current version", () => {
  // helm upgrade is the only way to update the values.yaml, so upgrade is
  // often used by users to update values only, so we can't default to the
  // latest version on the assumption that they always want to upgrade.
  const state = {
    ...defaultStore,
    apps: {
      selected: installedPkgDetail,
      selectedDetails: availablePkgDetail,
      isFetching: false,
    } as IInstalledPackageState,
    packages: {
      selected: selectedPkg,
    } as IPackageState,
  };

  const wrapper = mountWrapper(
    getStore({ ...state } as Partial<IStoreState>),
    <MemoryRouter initialEntries={[routePathParam]}>
      <Route path={routePath}>
        <UpgradeForm />,
      </Route>
    </MemoryRouter>,
  );
  expect(wrapper.find(DeploymentFormBody).prop("packageVersion")).toBe("1.0.0");
});

it("uses the selected version passed in the component's props", () => {
  const mockDispatch = jest.fn().mockReturnValue(true);
  jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
  const fetchAndSelectAvailablePackageDetail = jest.fn();
  actions.availablepackages.fetchAndSelectAvailablePackageDetail =
    fetchAndSelectAvailablePackageDetail;

  const state = {
    ...defaultStore,
    apps: {
      selected: installedPkgDetail,
      selectedDetails: availablePkgDetail,
      isFetching: false,
    } as IInstalledPackageState,
    packages: {
      selected: {
        ...selectedPkg,
        versions: [] as PackageAppVersion[],
      },
    } as IPackageState,
  };

  mountWrapper(
    getStore({ ...state } as Partial<IStoreState>),
    <MemoryRouter initialEntries={[routePathParam + "/1.5.0"]}>
      <Route path={routePath + "/:version"}>
        <UpgradeForm version={"1.5.0"} />,
      </Route>
    </MemoryRouter>,
  );

  expect(fetchAndSelectAvailablePackageDetail).toHaveBeenCalledWith(
    {
      context: {
        cluster: defaultProps.cluster,
        namespace: defaultProps.repoNamespace,
      },
      identifier: defaultProps.packageId,
      plugin: defaultProps.plugin,
    },
    "1.5.0",
  );
});

it("forwards the appValues when modified", () => {
  const state = {
    ...defaultStore,
    apps: {
      selected: { ...installedPkgDetail, valuesApplied: "foo: not-bar" },
      selectedDetails: { ...availablePkgDetail, defaultValues: "# A comment\nfoo: bar\n" },
      isFetching: false,
    } as IInstalledPackageState,
    packages: {
      selected: { ...selectedPkg, values: "initial: values" },
    } as IPackageState,
  };
  const wrapper = mountWrapper(
    getStore({ ...state } as Partial<IStoreState>),
    <MemoryRouter initialEntries={[routePathParam]}>
      <Route path={routePath}>
        <UpgradeForm />,
      </Route>
    </MemoryRouter>,
  );

  const handleValuesChange: (v: string) => void = wrapper
    .find(DeploymentFormBody)
    .prop("setValues");
  act(() => {
    handleValuesChange("foo: bar");
  });
  wrapper.update();

  expect(wrapper.find(DeploymentFormBody).prop("appValues")).toBe("foo: bar");
});

it("triggers an upgrade when submitting the form", async () => {
  const mockDispatch = jest.fn().mockReturnValue(true);
  jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
  const updateInstalledPackage = jest.fn();
  actions.installedpackages.updateInstalledPackage = updateInstalledPackage;

  const appValues = 'initial: values\nfoo: "bar"\n';
  const state = {
    ...defaultStore,
    apps: {
      selected: { ...installedPkgDetail, valuesApplied: appValues },
      selectedDetails: availablePkgDetail,
      isFetching: false,
    } as IInstalledPackageState,
    packages: {
      selected: selectedPkg,
    } as IPackageState,
  };

  const wrapper = mountWrapper(
    getStore({ ...state } as Partial<IStoreState>),
    <MemoryRouter initialEntries={[routePathParam]}>
      <Route path={routePath}>
        <UpgradeForm />,
      </Route>
    </MemoryRouter>,
  );

  await act(async () => {
    const handleValuesChange: (v: string) => void = wrapper
      .find(DeploymentFormBody)
      .prop("setValues");
    handleValuesChange(appValues);
    // Simulating "submit" causes a console.warning
    (wrapper.find("form").prop("onSubmit") as (e: any) => Promise<void>)({
      preventDefault: jest.fn(),
    });
  });
  expect(updateInstalledPackage).toHaveBeenCalledWith(
    installedPkgDetail.installedPackageRef,
    availablePkgDetail,
    appValues,
    schema,
  );
  expect(mockDispatch).toHaveBeenCalledWith({
    payload: {
      args: [url.app.apps.get(installedPkgDetail.installedPackageRef!)],
      method: "push",
    },
    type: "@@router/CALL_HISTORY_METHOD",
  });
});

describe("when receiving new props", () => {
  it("should calculate the modifications from the default and the current values", () => {
    const defaultValues = "initial: values\na: b\n";
    const deployedValues = "a: b\n";
    const currentValues = 'a: b\nc: "d"\n';
    const expectedValues = 'initial: values\na: b\nc: "d"\n';

    const state = {
      ...defaultStore,
      apps: {
        selected: { ...installedPkgDetail, valuesApplied: currentValues },
        selectedDetails: { ...availablePkgDetail, defaultValues: deployedValues },
        isFetching: false,
      } as IInstalledPackageState,
      packages: {
        selected: { ...selectedPkg, values: defaultValues },
      } as IPackageState,
    };

    const wrapper = mountWrapper(
      getStore({ ...state } as Partial<IStoreState>),
      <MemoryRouter initialEntries={[routePathParam]}>
        <Route path={routePath}>
          <UpgradeForm />,
        </Route>
      </MemoryRouter>,
    );

    wrapper.setProps({ deployed: { values: defaultValues } });
    expect(wrapper.find(DeploymentFormBody).prop("appValues")).toEqual(expectedValues);
  });

  it("should apply modifications if a new version is selected", () => {
    const defaultValues = "a: b\n";
    const deployedValues = "a: B\n";
    const currentValues = 'a: B\nc: "d"\n';
    const expectedValues = 'a: b\nc: "d"\n';
    const state = {
      ...defaultStore,
      apps: {
        selected: { ...installedPkgDetail, valuesApplied: currentValues },
        selectedDetails: { ...availablePkgDetail, defaultValues: deployedValues },
        isFetching: false,
      } as IInstalledPackageState,
      packages: {
        selected: { ...selectedPkg, values: defaultValues },
      } as IPackageState,
    };
    const wrapper = mountWrapper(
      getStore({ ...state } as Partial<IStoreState>),
      <MemoryRouter initialEntries={[routePathParam]}>
        <Route path={routePath}>
          <UpgradeForm />,
        </Route>
      </MemoryRouter>,
    );

    expect(wrapper.find(DeploymentFormBody).prop("appValues")).toEqual(expectedValues);
  });

  it("won't apply changes if the values have been manually modified", () => {
    const userValues = "a: b\n";
    const state = {
      ...defaultStore,
      apps: {
        selected: installedPkgDetail,
        selectedDetails: availablePkgDetail,
        isFetching: false,
      } as IInstalledPackageState,
      packages: {
        selected: selectedPkg,
      } as IPackageState,
    };
    const wrapper = mountWrapper(
      getStore({ ...state }),
      <MemoryRouter initialEntries={[routePathParam]}>
        <Route path={routePath}>
          <UpgradeForm />,
        </Route>
      </MemoryRouter>,
    );
    act(() => {
      const handleValuesChange: (v: string) => void = wrapper
        .find(DeploymentFormBody)
        .prop("setValues");
      handleValuesChange(userValues);
      const setValuesModified: () => void = wrapper
        .find(DeploymentFormBody)
        .prop("setValuesModified");
      setValuesModified();
    });
    wrapper.setProps({ selected: { versions: [testVersion], version: availablePkgDetail } });
    wrapper.update();
    expect(wrapper.find(DeploymentFormBody).prop("appValues")).toEqual(userValues);
  });

  [
    {
      description: "should merge modifications from the values and the new version defaults",
      defaultValues: "foo: bar\n",
      deployedValues: 'foo: bar\nmy: "var"\n',
      newDefaultValues: "notFoo: bar",
      result: 'notFoo: bar\nmy: "var"\n',
    },
    {
      description: "should modify the default values",
      defaultValues: "foo: bar\n",
      deployedValues: "foo: BAR\nmy: var\n",
      newDefaultValues: "foo: bar",
      result: 'foo: BAR\nmy: "var"\n',
    },
    {
      description: "should delete an element in the defaults",
      defaultValues: "foo: bar\n",
      deployedValues: "my: var\n",
      newDefaultValues: "foo: bar\n",
      result: 'my: "var"\n',
    },
    {
      description: "should add an element in an array",
      defaultValues: `foo:
  - foo1:
    bar1: value1
`,
      deployedValues: `foo:
  - foo1:
    bar1: value1
  - foo2:
    bar2: value2
`,
      newDefaultValues: `foo:
    - foo1:
      bar1: value1
`,
      result: [`foo:`, `  - foo1:`, `    bar1: value1`, `  - foo2:`, `    bar2: "value2"`, ``].join(
        "\n",
      ),
    },
    {
      description: "should delete an element in an array",
      defaultValues: `foo:
  - foo1:
    bar1: value1
  - foo2:
    bar2: value2
`,
      deployedValues: `foo:
  - foo1:
    bar1: value1
`,
      newDefaultValues: `foo:
  - foo1:
    bar1: value1
  - foo2:
    bar2: value2
`,
      result: [`foo:`, `  - foo1:`, `    bar1: value1`, ``].join("\n"),
    },
    {
      description: "set a value with dots and slashes in the key",
      defaultValues: "foo.bar/foobar: ",
      deployedValues: "foo.bar/foobar: value",
      newDefaultValues: "foo.bar/foobar: ",
      result: "foo.bar/foobar: value\n",
    },
  ].forEach(t => {
    it(t.description, () => {
      const state = {
        ...defaultStore,
        apps: {
          selected: { ...installedPkgDetail, valuesApplied: t.deployedValues },
          selectedDetails: { ...availablePkgDetail, defaultValues: t.defaultValues },
          isFetching: false,
        } as IInstalledPackageState,
        packages: {
          selected: { ...selectedPkg, values: "initial: values" },
        } as IPackageState,
      };
      const newState = {
        ...state,
        apps: {
          ...state.apps,
          selected: {
            ...state.apps.selected,
            values: t.deployedValues,
          },
        } as IInstalledPackageState,
        packages: {
          selected: {
            ...state.packages.selected,
            values: t.newDefaultValues,
          },
        } as IPackageState,
      };

      const wrapper = mountWrapper(
        getStore({ ...newState }),
        <MemoryRouter initialEntries={[routePathParam]}>
          <Route path={routePath}>
            <UpgradeForm />,
          </Route>
        </MemoryRouter>,
      );
      expect(wrapper.find(DeploymentFormBody).prop("appValues")).toEqual(t.result);
    });
  });
});

it("shows, by default, the default values of the deployed package plus any modification", () => {
  const defaultValues = "initial: values";
  const deployedValues = "# A comment\nfoo: bar\n";
  const currentValues = "foo: not-bar";
  const expectedValues = "# A comment\nfoo: not-bar\n";

  const state = {
    ...defaultStore,
    apps: {
      selected: { ...installedPkgDetail, valuesApplied: currentValues },
      selectedDetails: { ...availablePkgDetail, defaultValues: deployedValues },
      isFetching: false,
    } as IInstalledPackageState,
    packages: {
      selected: { ...selectedPkg, values: defaultValues },
    } as IPackageState,
  };
  const wrapper = mountWrapper(
    getStore({ ...state } as Partial<IStoreState>),
    <MemoryRouter initialEntries={[routePathParam]}>
      <Route path={routePath}>
        <UpgradeForm />,
      </Route>
    </MemoryRouter>,
  );

  expect(wrapper.find(DeploymentFormBody).prop("deployedValues")).toBe(expectedValues);
});
