// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { deepClone } from "@cds/core/internal";
import actions from "actions";
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
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import context from "jest-plugin-context";
import qs from "qs";
import React from "react";
import { act } from "react-dom/test-utils";
import * as ReactRedux from "react-redux";
import { MemoryRouter } from "react-router-dom";
import { Kube } from "shared/Kube";
import { defaultStore, getStore, initialState, mountWrapper } from "shared/specs/mountWrapper";
import { FetchError, IStoreState } from "shared/types";
import Alert from "../js/Alert";
import AppList from "./AppList";
import AppListItem from "./AppListItem";
import CustomResourceListItem from "./CustomResourceListItem";

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
    state.config.featureFlags = { operators: true };
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
    state.config.featureFlags = { operators: false };
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
    );
    expect(wrapper.find(SearchFilter).prop("value")).toEqual("foo");
  });

  it("should list apps in all namespaces", () => {
    jest.spyOn(qs, "parse").mockReturnValue({
      allns: "yes",
    });
    const wrapper = mountWrapper(
      defaultStore,
      <MemoryRouter initialEntries={["/foo?allns=yes"]}>
        <AppList />
      </MemoryRouter>,
    );
    expect(wrapper.find("input[type='checkbox']")).toBeChecked();
  });

  it("should fetch apps in all namespaces", async () => {
    const state = deepClone(initialState) as IStoreState;
    state.config.featureFlags = { operators: true };
    const store = getStore(state);
    const fetchInstalledPackages = jest.fn();
    const getCustomResources = jest.fn();
    actions.installedpackages.fetchInstalledPackages = fetchInstalledPackages;
    actions.operators.getResources = getCustomResources;
    const wrapper = mountWrapper(store, <AppList />);
    act(() => {
      wrapper.find("input[type='checkbox']").simulate("change");
    });
    expect(fetchInstalledPackages).toHaveBeenCalledWith("default-cluster", "");
    expect(getCustomResources).toHaveBeenCalledWith("default-cluster", "");
  });

  it("should hide the all-namespace switch if the user doesn't have permissions", async () => {
    Kube.canI = jest.fn().mockReturnValue({
      then: jest.fn((f: any) => f(false)),
      catch: jest.fn(f => f(false)),
    });
    const wrapper = mountWrapper(defaultStore, <AppList />);
    expect(wrapper.find("input[type='checkbox']")).not.toExist();
  });

  describe("when store changes", () => {
    let spyOnUseState: jest.SpyInstance;
    afterEach(() => {
      spyOnUseState.mockRestore();
    });

    it("should not set all-ns prop when getting changes in the namespace", async () => {
      const setAllNS = jest.fn();
      const useState = jest.fn();
      spyOnUseState = jest
        .spyOn(React, "useState")
        /* @ts-expect-error: Argument of type '(init: any) => any[]' is not assignable to parameter of type '() => [unknown, Dispatch<unknown>]' */
        .mockImplementation((init: any) => {
          if (init === false) {
            // Mocking the result of setAllNS
            return [false, setAllNS];
          }
          return [init, useState];
        });

      mountWrapper(
        defaultStore,
        <MemoryRouter initialEntries={["/foo?allns=yes"]}>
          <AppList />
        </MemoryRouter>,
      );
      expect(setAllNS).not.toHaveBeenCalledWith(false);
    });
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
    expect(wrapper.find(Alert)).toExist();
    expect(wrapper.find(Alert).html()).toContain("Boom!");
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
        installedPackageRef: {
          identifier: "bar/foo",
          pkgVersion: "1.0.0",
          context: { cluster: "", namespace: "foobar" } as Context,
          plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
        } as InstalledPackageReference,
        status: {
          ready: true,
          reason: InstalledPackageStatus_StatusReason.STATUS_REASON_INSTALLED,
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
        installedPackageRef: {
          identifier: "foo/bar",
          pkgVersion: "1.0.0",
          context: { cluster: "", namespace: "fooNs" } as Context,
          plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
        } as InstalledPackageReference,
        status: {
          ready: true,
          reason: InstalledPackageStatus_StatusReason.STATUS_REASON_INSTALLED,
          userReason: "deployed",
        } as InstalledPackageStatus,
        latestMatchingVersion: { appVersion: "0.1.0", pkgVersion: "1.0.0" } as PackageAppVersion,
        latestVersion: { appVersion: "0.1.0", pkgVersion: "1.0.0" } as PackageAppVersion,
        currentVersion: { appVersion: "0.1.0", pkgVersion: "1.0.0" } as PackageAppVersion,
        pkgVersionReference: { version: "1" } as VersionReference,
      } as InstalledPackageSummary,
      {
        name: "bar",
        installedPackageRef: {
          identifier: "foobar/bar",
          pkgVersion: "1.0.0",
          context: { cluster: "", namespace: "fooNs" } as Context,
          plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
        } as InstalledPackageReference,
        status: {
          ready: true,
          reason: InstalledPackageStatus_StatusReason.STATUS_REASON_INSTALLED,
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
    );
    expect(wrapper.find(AppListItem).key()).toBe("fooNs-foobar/bar");
  });

  it("filters apps (same name, different ns)", () => {
    state.apps.listOverview = [
      {
        name: "foo",
        installedPackageRef: {
          identifier: "foo/bar",
          pkgVersion: "1.0.0",
          context: { cluster: "", namespace: "fooNs" } as Context,
          plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
        } as InstalledPackageReference,
        status: {
          ready: true,
          reason: InstalledPackageStatus_StatusReason.STATUS_REASON_INSTALLED,
          userReason: "deployed",
        } as InstalledPackageStatus,
        latestMatchingVersion: { appVersion: "0.1.0", pkgVersion: "1.0.0" } as PackageAppVersion,
        latestVersion: { appVersion: "0.1.0", pkgVersion: "1.0.0" } as PackageAppVersion,
        currentVersion: { appVersion: "0.1.0", pkgVersion: "1.0.0" } as PackageAppVersion,
        pkgVersionReference: { version: "1" } as VersionReference,
      } as InstalledPackageSummary,
      {
        name: "bar",
        installedPackageRef: {
          identifier: "foobar/bar",
          pkgVersion: "1.0.0",
          context: { cluster: "", namespace: "fooNs" } as Context,
          plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
        } as InstalledPackageReference,
        status: {
          ready: true,
          reason: InstalledPackageStatus_StatusReason.STATUS_REASON_INSTALLED,
          userReason: "deployed",
        } as InstalledPackageStatus,
        latestMatchingVersion: { appVersion: "0.1.0", pkgVersion: "1.0.0" } as PackageAppVersion,
        latestVersion: { appVersion: "0.1.0", pkgVersion: "1.0.0" } as PackageAppVersion,
        currentVersion: { appVersion: "0.1.0", pkgVersion: "1.0.0" } as PackageAppVersion,
        pkgVersionReference: { version: "1" } as VersionReference,
      } as InstalledPackageSummary,
      {
        name: "bar",
        installedPackageRef: {
          identifier: "foobar/bar",
          pkgVersion: "1.0.0",
          context: { cluster: "", namespace: "barNs" } as Context,
          plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
        } as InstalledPackageReference,
        status: {
          ready: true,
          reason: InstalledPackageStatus_StatusReason.STATUS_REASON_INSTALLED,
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
    );
    const itemList = wrapper.find(CustomResourceListItem);
    expect(itemList).not.toExist();
  });
});
