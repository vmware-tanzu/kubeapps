// Copyright 2018-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { deepClone } from "@cds/core/internal";
import actions from "actions";
import AlertGroup from "components/AlertGroup";
import LoadingWrapper from "components/LoadingWrapper";
import SearchFilter from "components/SearchFilter/SearchFilter";
import {
  Context,
  InstalledPackageReference,
  InstalledPackageStatus,
  InstalledPackageStatus_StatusReason,
  InstalledPackageSummary,
  PackageAppVersion,
  VersionReference,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages_pb";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins_pb";
import context from "jest-plugin-context";
import qs from "qs";
import * as ReactRedux from "react-redux";
import { MemoryRouter } from "react-router-dom";
import { Kube } from "shared/Kube";
import {
  defaultStore,
  getStore,
  initialState,
  mountWrapper,
  renderWithProviders,
} from "shared/specs/mountWrapper";
import { FetchError, IStoreState } from "shared/types";
import AppList from "./AppList";
import AppListItem from "./AppListItem";
import CustomResourceListItem from "./CustomResourceListItem";
import { screen } from "@testing-library/react";

let spyOnUseDispatch: jest.SpyInstance;
const opActions = { ...actions.operators };
const appActions = { ...actions.installedpackages };
beforeEach(() => {
  actions.operators = {
    ...actions.operators,
    getResources: jest.fn(),
  };
  actions.installedpackages = {
    ...actions.installedpackages,
    fetchInstalledPackages: jest.fn(),
  };
  const mockDispatch = jest.fn();
  spyOnUseDispatch = jest.spyOn(ReactRedux, "useDispatch").mockReturnValue(mockDispatch);
  Kube.canI = jest.fn().mockReturnValue({
    then: jest.fn(f => f(true)),
    catch: jest.fn(f => f(false)),
  });
});

afterEach(() => {
  actions.operators = { ...opActions };
  actions.installedpackages = { ...appActions };
  spyOnUseDispatch.mockRestore();
});

context("when changing props", () => {
  it("should fetch apps in the new namespace", async () => {
    const state = deepClone(initialState) as IStoreState;
    state.config.featureFlags = { ...initialState.config.featureFlags, operators: true };
    const store = getStore(state);
    const fetchInstalledPackages = jest.fn();
    const getCustomResources = jest.fn();
    actions.installedpackages.fetchInstalledPackages = fetchInstalledPackages;
    actions.operators.getResources = getCustomResources;
    mountWrapper(store, <AppList />);
    expect(fetchInstalledPackages).toHaveBeenCalledWith("default-cluster", "default");
    expect(getCustomResources).toHaveBeenCalledWith("default-cluster", "default");
  });

  it("should not fetch resources in the new namespace when operators is deactivated", async () => {
    const state = deepClone(initialState) as IStoreState;
    state.config.featureFlags = { ...initialState.config.featureFlags, operators: false };
    const store = getStore(state);
    const fetchInstalledPackages = jest.fn();
    const getCustomResources = jest.fn();
    actions.installedpackages.fetchInstalledPackages = fetchInstalledPackages;
    actions.operators.getResources = getCustomResources;
    mountWrapper(store, <AppList />);
    expect(fetchInstalledPackages).toHaveBeenCalledWith("default-cluster", "default");
    expect(getCustomResources).not.toHaveBeenCalled();
  });

  it("should update the search filter", () => {
    jest.spyOn(qs, "parse").mockReturnValue({
      q: "foo",
    });
    const wrapper = mountWrapper(
      defaultStore,
      <MemoryRouter initialEntries={["/foo?q=foo"]}>
        <AppList />
      </MemoryRouter>,
      false,
    );
    expect(wrapper.find(SearchFilter).prop("value")).toEqual("foo");
  });

  it("should fetch apps in all namespaces", async () => {
    const fetchInstalledPackages = jest.fn();
    const getCustomResources = jest.fn();
    actions.installedpackages.fetchInstalledPackages = fetchInstalledPackages;
    actions.operators.getResources = getCustomResources;

    renderWithProviders(<AppList />, {
      preloadedState: {
        clusters: {
          currentCluster: "default-cluster",
          clusters: {
            "default-cluster": {
              currentNamespace: "default",
            },
          },
        },
        config: {
          featureFlags: {
            operators: true,
          },
        },
      },
      initialEntries: ["/c/default-cluster/ns/default/apps?allns=no"],
    });

    expect(fetchInstalledPackages).toHaveBeenCalledTimes(1);
    expect(fetchInstalledPackages).toHaveBeenCalledWith("default-cluster", "default");
    expect(getCustomResources).toHaveBeenCalledTimes(1);
    expect(getCustomResources).toHaveBeenCalledWith("default-cluster", "default");

    screen.getByRole("checkbox").click();

    expect(fetchInstalledPackages).toHaveBeenCalledTimes(2);
    expect(fetchInstalledPackages).toHaveBeenCalledWith("default-cluster", "");
    expect(getCustomResources).toHaveBeenCalledTimes(2);
    expect(getCustomResources).toHaveBeenCalledWith("default-cluster", "");
  });

  it("should not requests apps if namespace not set", async () => {
    // If a page is reloaded, the namespace is not yet set in the state, so sending
    // off a request at that point returns apps for all namespaces.
    const fetchInstalledPackages = jest.fn();
    const getCustomResources = jest.fn();
    actions.installedpackages.fetchInstalledPackages = fetchInstalledPackages;
    actions.operators.getResources = getCustomResources;

    renderWithProviders(<AppList />, {
      preloadedState: {
        clusters: {
          currentCluster: "default-cluster",
          clusters: {
            "default-cluster": {
              currentNamespace: "",
            },
          },
        },
        config: {
          featureFlags: {
            operators: true,
          },
        },
      },
      initialEntries: ["/c/default-cluster/ns/default/apps?allns=no"],
    });

    expect(fetchInstalledPackages).toHaveBeenCalledTimes(0);
    expect(getCustomResources).toHaveBeenCalledTimes(0);
  });

  it("should hide the all-namespace switch if the user doesn't have permissions", async () => {
    Kube.canI = jest.fn().mockReturnValue({
      then: jest.fn((f: any) => f(false)),
      catch: jest.fn(f => f(false)),
    });
    const wrapper = mountWrapper(defaultStore, <AppList />);
    expect(wrapper.find("input[type='checkbox']")).not.toExist();
  });
});

context("while fetching apps", () => {
  const state = deepClone(initialState) as IStoreState;
  state.apps.isFetching = true;

  it("behaves as loading component", () => {
    const wrapper = mountWrapper(getStore(state), <AppList />);
    expect(wrapper.find(LoadingWrapper)).toExist();
  });

  it("behaves as loading component (for operators)", () => {
    const stateOp = deepClone(initialState) as IStoreState;
    state.operators.isFetching = true;
    const wrapper = mountWrapper(getStore(stateOp), <AppList />);
    expect(wrapper.find(LoadingWrapper)).toExist();
  });

  it("renders a Application header (while fetching)", () => {
    const wrapper = mountWrapper(getStore(state), <AppList />);
    expect(wrapper.find("h1").text()).toContain("Applications");
  });

  it("shows the search filter and deploy button (while fetching)", () => {
    const wrapper = mountWrapper(getStore(state), <AppList />);
    expect(wrapper.find("SearchFilter")).toExist();
    expect(wrapper.find("Link")).toExist();
  });
});

context("when fetched but not apps available", () => {
  it("renders a welcome message", () => {
    const wrapper = mountWrapper(defaultStore, <AppList />);
    expect(wrapper.find(".applist-empty").text()).toContain("Welcome To Kubeapps");
  });

  it("shows the search filter and deploy button (no apps available)", () => {
    const wrapper = mountWrapper(defaultStore, <AppList />);
    expect(wrapper.find("SearchFilter")).toExist();
    expect(wrapper.find("Link")).toExist();
  });
});

context("when an error is present", () => {
  const state = deepClone(initialState) as IStoreState;
  state.apps.error = new FetchError("Boom!");

  it("renders a generic error message", () => {
    const wrapper = mountWrapper(getStore(state), <AppList />);
    expect(wrapper.find(AlertGroup)).toExist();
    expect(wrapper.find(AlertGroup).html()).toContain("Boom!");
  });

  it("renders a Application header (when error)", () => {
    const wrapper = mountWrapper(getStore(state), <AppList />);
    expect(wrapper.find("h1").text()).toContain("Applications");
  });
});

context("when apps available", () => {
  const state = deepClone(initialState) as IStoreState;
  beforeEach(() => {
    state.apps.listOverview = [
      {
        name: "foo",
        installedPackageRef: new InstalledPackageReference({
          identifier: "bar/foo",
          context: { cluster: "", namespace: "foobar" } as Context,
          plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
        }),
        status: {
          ready: true,
          reason: InstalledPackageStatus_StatusReason.INSTALLED,
          userReason: "deployed",
        } as InstalledPackageStatus,
        latestMatchingVersion: { appVersion: "0.1.0", pkgVersion: "1.0.0" } as PackageAppVersion,
        latestVersion: { appVersion: "0.1.0", pkgVersion: "1.0.0" } as PackageAppVersion,
        currentVersion: { appVersion: "0.1.0", pkgVersion: "1.0.0" } as PackageAppVersion,
        pkgVersionReference: { version: "1" } as VersionReference,
      } as InstalledPackageSummary,
    ];
  });
  afterEach(() => {
    jest.restoreAllMocks();
  });

  it("renders a CardGrid with the available Apps", () => {
    const wrapper = mountWrapper(getStore(state), <AppList />);
    const itemList = wrapper.find(AppListItem);
    expect(itemList).toExist();
    expect(itemList.key()).toBe("foobar-bar/foo");
  });

  it("filters apps", () => {
    state.apps.listOverview = [
      {
        name: "foo",
        installedPackageRef: new InstalledPackageReference({
          identifier: "foo/bar",
          context: { cluster: "", namespace: "fooNs" } as Context,
          plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
        }),
        status: {
          ready: true,
          reason: InstalledPackageStatus_StatusReason.INSTALLED,
          userReason: "deployed",
        } as InstalledPackageStatus,
        latestMatchingVersion: { appVersion: "0.1.0", pkgVersion: "1.0.0" } as PackageAppVersion,
        latestVersion: { appVersion: "0.1.0", pkgVersion: "1.0.0" } as PackageAppVersion,
        currentVersion: { appVersion: "0.1.0", pkgVersion: "1.0.0" } as PackageAppVersion,
        pkgVersionReference: { version: "1" } as VersionReference,
      } as InstalledPackageSummary,
      {
        name: "bar",
        installedPackageRef: new InstalledPackageReference({
          identifier: "foobar/bar",
          context: { cluster: "", namespace: "fooNs" } as Context,
          plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
        }),
        status: {
          ready: true,
          reason: InstalledPackageStatus_StatusReason.INSTALLED,
          userReason: "deployed",
        } as InstalledPackageStatus,
        latestMatchingVersion: { appVersion: "0.1.0", pkgVersion: "1.0.0" } as PackageAppVersion,
        latestVersion: { appVersion: "0.1.0", pkgVersion: "1.0.0" } as PackageAppVersion,
        currentVersion: { appVersion: "0.1.0", pkgVersion: "1.0.0" } as PackageAppVersion,
        pkgVersionReference: { version: "1" } as VersionReference,
      } as InstalledPackageSummary,
    ];
    jest.spyOn(qs, "parse").mockReturnValue({
      q: "bar",
    });
    const wrapper = mountWrapper(
      getStore(state),
      <MemoryRouter initialEntries={["/foo?q=bar"]}>
        <AppList />
      </MemoryRouter>,
      false,
    );
    expect(wrapper.find(AppListItem).key()).toBe("fooNs-foobar/bar");
  });

  it("filters apps (same name, different ns)", () => {
    state.apps.listOverview = [
      {
        name: "foo",
        installedPackageRef: new InstalledPackageReference({
          identifier: "foo/bar",
          context: { cluster: "", namespace: "fooNs" } as Context,
          plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
        }),
        status: {
          ready: true,
          reason: InstalledPackageStatus_StatusReason.INSTALLED,
          userReason: "deployed",
        } as InstalledPackageStatus,
        latestMatchingVersion: { appVersion: "0.1.0", pkgVersion: "1.0.0" } as PackageAppVersion,
        latestVersion: { appVersion: "0.1.0", pkgVersion: "1.0.0" } as PackageAppVersion,
        currentVersion: { appVersion: "0.1.0", pkgVersion: "1.0.0" } as PackageAppVersion,
        pkgVersionReference: { version: "1" } as VersionReference,
      } as InstalledPackageSummary,
      {
        name: "bar",
        installedPackageRef: new InstalledPackageReference({
          identifier: "foobar/bar",
          context: { cluster: "", namespace: "fooNs" } as Context,
          plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
        }),
        status: {
          ready: true,
          reason: InstalledPackageStatus_StatusReason.INSTALLED,
          userReason: "deployed",
        } as InstalledPackageStatus,
        latestMatchingVersion: { appVersion: "0.1.0", pkgVersion: "1.0.0" } as PackageAppVersion,
        latestVersion: { appVersion: "0.1.0", pkgVersion: "1.0.0" } as PackageAppVersion,
        currentVersion: { appVersion: "0.1.0", pkgVersion: "1.0.0" } as PackageAppVersion,
        pkgVersionReference: { version: "1" } as VersionReference,
      } as InstalledPackageSummary,
      {
        name: "bar",
        installedPackageRef: new InstalledPackageReference({
          identifier: "foobar/bar",
          context: { cluster: "", namespace: "barNs" } as Context,
          plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
        }),
        status: {
          ready: true,
          reason: InstalledPackageStatus_StatusReason.INSTALLED,
          userReason: "deployed",
        } as InstalledPackageStatus,
        latestMatchingVersion: { appVersion: "0.1.0", pkgVersion: "1.0.0" } as PackageAppVersion,
        latestVersion: { appVersion: "0.1.0", pkgVersion: "1.0.0" } as PackageAppVersion,
        currentVersion: { appVersion: "0.1.0", pkgVersion: "1.0.0" } as PackageAppVersion,
        pkgVersionReference: { version: "1" } as VersionReference,
      } as InstalledPackageSummary,
    ];
    jest.spyOn(qs, "parse").mockReturnValue({
      q: "bar",
    });
    const wrapper = mountWrapper(
      getStore(state),
      <MemoryRouter initialEntries={["/foo?q=bar"]}>
        <AppList />
      </MemoryRouter>,
      false,
    );
    expect(wrapper.find(AppListItem).first().key()).toBe("fooNs-foobar/bar");
    expect(wrapper.find(AppListItem).last().key()).toBe("barNs-foobar/bar");
  });
});

context("when custom resources available", () => {
  const state = deepClone(initialState) as IStoreState;
  const cr = {
    kind: "KubeappsCluster",
    metadata: { name: "foo-cluster", namespace: "foo-ns" },
  } as any;
  const csv = {
    metadata: {
      name: "foo",
      namespace: "foo-ns",
    },
    spec: {
      customresourcedefinitions: {
        owned: [
          {
            kind: "KubeappsCluster",
          },
        ],
      },
    },
  } as any;

  beforeEach(() => {
    state.operators.resources = [cr];
    state.operators.csvs = [csv];
  });
  afterEach(() => {
    jest.restoreAllMocks();
  });

  it("renders a CardGrid with the available resources", () => {
    const wrapper = mountWrapper(getStore(state), <AppList />);
    const itemList = wrapper.find(CustomResourceListItem);
    expect(itemList).toExist();
    expect(itemList.key()).toBe("foo-cluster_foo-ns");
  });

  it("filters out items", () => {
    jest.spyOn(qs, "parse").mockReturnValue({
      q: "nop",
    });
    const wrapper = mountWrapper(
      getStore(state),
      <MemoryRouter initialEntries={["/foo?q=nop"]}>
        <AppList />
      </MemoryRouter>,
      false,
    );
    const itemList = wrapper.find(CustomResourceListItem);
    expect(itemList).not.toExist();
  });
});
