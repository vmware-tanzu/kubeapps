// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import actions from "actions";
import Alert from "components/js/Alert";
import LoadingWrapper from "components/LoadingWrapper";
import {
  AvailablePackageDetail,
  AvailablePackageReference,
  Context,
  InstalledPackageReference,
  InstalledPackageStatus,
  InstalledPackageStatus_StatusReason,
  PackageAppVersion,
  VersionReference,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import {
  PackageRepositoryDetail,
  PackageRepositorySummary,
} from "gen/kubeappsapis/core/packages/v1alpha1/repositories";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import * as ReactRedux from "react-redux";
import * as ReactRouter from "react-router";
import { MemoryRouter, Route } from "react-router-dom";
import { IPackageRepositoryState } from "reducers/repos";
import { defaultStore, getStore, mountWrapper } from "shared/specs/mountWrapper";
import {
  CustomInstalledPackageDetail,
  FetchError,
  IInstalledPackageState,
  IPackageState,
  IStoreState,
  PluginNames,
  UpgradeError,
} from "shared/types";
import SelectRepoForm from "../SelectRepoForm/SelectRepoForm";
import UpgradeForm from "../UpgradeForm/UpgradeForm";
import AppUpgrade from "./AppUpgrade";

const defaultProps = {
  pkgName: "foo",
  cluster: "default",
  namespace: "default",
  repoNamespace: "stable",
  repo: "repo",
  releaseName: "my-release",
  plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
};

const installedPackage1 = {
  name: "test",
  postInstallationNotes: "test",
  valuesApplied: "test",
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

const availablePackageDetail = {
  availablePackageRef: {
    context: { cluster: "default", namespace: "my-ns" },
    identifier: "test",
    plugin: { name: PluginNames.PACKAGES_HELM, version: "0.0.1" } as Plugin,
  },
  version: { appVersion: "4.5.6", pkgVersion: "1.2.3" },
} as AvailablePackageDetail;

const selectedPackage = {
  versions: [{ appVersion: "10.0.0", pkgVersion: "1.2.3" }],
  availablePackageDetail: { name: "test" } as AvailablePackageDetail,
} as IPackageState["selected"];

const repo1Summary = {
  name: defaultProps.repo,
  packageRepoRef: {
    context: { namespace: defaultProps.repoNamespace, cluster: defaultProps.cluster },
    identifier: defaultProps.repo,
    plugin: defaultProps.plugin,
  },
} as PackageRepositorySummary;

const repo1Detail = {
  name: defaultProps.repo,
  packageRepoRef: {
    context: { namespace: defaultProps.repoNamespace, cluster: defaultProps.cluster },
    identifier: defaultProps.repo,
    plugin: defaultProps.plugin,
  },
} as PackageRepositoryDetail;

let spyOnUseDispatch: jest.SpyInstance;
let spyOnUseHistory: jest.SpyInstance;

beforeEach(() => {
  const mockDispatch = jest.fn();
  spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
  spyOnUseHistory = jest
    .spyOn(ReactRouter, "useHistory")
    .mockReturnValue({ push: jest.fn() } as any);
});

afterEach(() => {
  jest.restoreAllMocks();
  spyOnUseDispatch.mockRestore();
  spyOnUseHistory.mockRestore();
});

const routePathParam = `/c/${defaultProps.cluster}/ns/${defaultProps.namespace}/apps/${defaultProps.plugin.name}/${defaultProps.plugin.version}/${defaultProps.releaseName}/upgrade`;
const routePath = "/c/:cluster/ns/:namespace/apps/:pluginName/:pluginVersion/:releaseName/upgrade";

it("renders the repo selection form if not introduced", () => {
  const state = {
    apps: {
      isFetching: true,
    } as IInstalledPackageState,
  };
  const wrapper = mountWrapper(
    getStore({ ...defaultStore, ...state } as Partial<IStoreState>),
    <MemoryRouter initialEntries={[routePathParam]}>
      <Route path={routePath}>
        <AppUpgrade />,
      </Route>
    </MemoryRouter>,
  );
  expect(wrapper.find(LoadingWrapper).prop("loaded")).toBe(false);
});

it("renders the repo selection form if not introduced when the app is loaded", () => {
  const state = {
    repos: {
      reposSummaries: [repo1Summary],
    } as IPackageRepositoryState,
    apps: {
      selected: { name: "foo" },
      isFetching: false,
      error: undefined,
    } as IInstalledPackageState,
  };
  const wrapper = mountWrapper(
    getStore({
      ...defaultStore,
      ...state,
    } as Partial<IStoreState>),
    <MemoryRouter initialEntries={[routePathParam]}>
      <Route path={routePath}>
        <AppUpgrade />,
      </Route>
    </MemoryRouter>,
  );
  expect(wrapper.find(SelectRepoForm)).toExist();
  expect(wrapper.find(Alert)).not.toExist();
  expect(wrapper.find(UpgradeForm)).not.toExist();
});

describe("when an error exists", () => {
  it("renders a generic error message", () => {
    const state = {
      apps: {
        error: new FetchError("foo does not exist"),
      } as IInstalledPackageState,
    };
    const wrapper = mountWrapper(
      getStore({
        ...defaultStore,
        ...state,
      } as Partial<IStoreState>),
      <MemoryRouter initialEntries={[routePathParam]}>
        <Route path={routePath}>
          <AppUpgrade />,
        </Route>
      </MemoryRouter>,
    );

    expect(wrapper.find(Alert)).toExist();
    expect(wrapper.find(SelectRepoForm)).not.toExist();
    expect(wrapper.find(UpgradeForm)).not.toExist();

    expect(wrapper.html()).toContain("foo does not exist");
  });

  it("renders a warning message if there are no repositories", () => {
    const state = {
      repos: {
        reposSummaries: [] as PackageRepositorySummary[],
      } as IPackageRepositoryState,
      apps: {
        selected: { name: "foo" },
        isFetching: false,
        error: undefined,
      } as IInstalledPackageState,
    };
    const wrapper = mountWrapper(
      getStore({
        ...defaultStore,
        ...state,
      } as Partial<IStoreState>),
      <MemoryRouter initialEntries={[routePathParam]}>
        <Route path={routePath}>
          <AppUpgrade />,
        </Route>
      </MemoryRouter>,
    );
    expect(wrapper.find(SelectRepoForm).find(Alert)).toExist();
    expect(wrapper.find(UpgradeForm)).not.toExist();

    expect(wrapper.find(Alert).children().text()).toContain("Repositories not found");
  });

  it("still renders the upgrade form even if there is an upgrade error", () => {
    const upgradeError = new UpgradeError("foo upgrade failed");
    const state = {
      apps: {
        error: upgradeError,
        selected: installedPackage1,
        selectedDetails: availablePackageDetail,
      } as IInstalledPackageState,
      packages: { selected: selectedPackage } as IPackageState,
    };

    const wrapper = mountWrapper(
      getStore({
        ...defaultStore,
        ...state,
      } as Partial<IStoreState>),
      <MemoryRouter initialEntries={[routePathParam]}>
        <Route path={routePath}>
          <AppUpgrade />,
        </Route>
      </MemoryRouter>,
    );
    expect(wrapper.find(UpgradeForm)).toExist();
    expect(wrapper.find(UpgradeForm).find(Alert)).toIncludeText(upgradeError.message);
  });
});

it("renders the upgrade form when the repo is available, clears state and fetches app", () => {
  const getInstalledPackage = jest.fn();
  actions.installedpackages.getInstalledPackage = getInstalledPackage;
  const resetSelectedAvailablePackageDetail = jest
    .spyOn(actions.availablepackages, "resetSelectedAvailablePackageDetail")
    .mockImplementation(jest.fn());

  const state = {
    apps: {
      selected: installedPackage1,
      selectedDetails: availablePackageDetail,
    } as IInstalledPackageState,
    repos: {
      repoDetail: repo1Detail,
      reposSummaries: [repo1Summary],
      isFetching: false,
    } as IPackageRepositoryState,
    packages: { selected: selectedPackage } as IPackageState,
  };
  const wrapper = mountWrapper(
    getStore({
      ...defaultStore,
      ...state,
    } as Partial<IStoreState>),
    <MemoryRouter initialEntries={[routePathParam]}>
      <Route path={routePath}>
        <AppUpgrade />,
      </Route>
    </MemoryRouter>,
  );
  expect(wrapper.find(UpgradeForm)).toExist();
  expect(wrapper.find(Alert)).not.toExist();
  expect(wrapper.find(SelectRepoForm)).not.toExist();

  expect(resetSelectedAvailablePackageDetail).toHaveBeenCalled();
  expect(getInstalledPackage).toHaveBeenCalledWith({
    context: { cluster: defaultProps.cluster, namespace: defaultProps.namespace },
    identifier: defaultProps.releaseName,
    plugin: defaultProps.plugin,
  });
});

it("renders the upgrade form with the version property", () => {
  const state = {
    apps: {
      selected: installedPackage1,
      selectedDetails: availablePackageDetail,
    } as IInstalledPackageState,
    repos: {
      repoDetail: repo1Detail,
      reposSummaries: [repo1Summary],
      isFetching: false,
    } as Partial<IPackageRepositoryState>,
    packages: { selected: selectedPackage } as IPackageState,
  };
  const wrapper = mountWrapper(
    getStore({
      ...defaultStore,
      ...state,
    } as Partial<IStoreState>),
    <MemoryRouter initialEntries={[routePathParam + "/0.0.1"]}>
      <Route path={routePath + "/:version"}>
        <AppUpgrade />,
      </Route>
    </MemoryRouter>,
  );
  expect(wrapper.find(UpgradeForm)).toExist();
  expect(wrapper.find(UpgradeForm)).toHaveProp("version", "0.0.1");
});

it("skips the repo selection form if the app contains upgrade info", () => {
  const state = {
    apps: {
      selected: installedPackage1,
      selectedDetails: availablePackageDetail,
    } as IInstalledPackageState,
    repos: {
      repoDetail: repo1Detail,
      reposSummaries: [repo1Summary],
      isFetching: false,
    } as IPackageRepositoryState,
    packages: { selected: selectedPackage } as IPackageState,
  };
  const wrapper = mountWrapper(
    getStore({
      ...defaultStore,
      ...state,
    } as Partial<IStoreState>),
    <MemoryRouter initialEntries={[routePathParam]}>
      <Route path={routePath}>
        <AppUpgrade />,
      </Route>
    </MemoryRouter>,
  );
  expect(wrapper.find(UpgradeForm)).toExist();
  expect(wrapper.find(Alert)).not.toExist();
  expect(wrapper.find(SelectRepoForm)).not.toExist();
});
