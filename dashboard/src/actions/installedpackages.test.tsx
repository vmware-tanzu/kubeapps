// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import {
  AvailablePackageDetail,
  Context,
  GetInstalledPackageSummariesResponse,
  InstalledPackageDetail,
  InstalledPackageReference,
  InstalledPackageStatus,
  InstalledPackageStatus_StatusReason,
  InstalledPackageSummary,
  ReconciliationOptions,
  VersionReference,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages_pb";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins_pb";
import { InstalledPackage } from "shared/InstalledPackage";
import { getStore, initialState } from "shared/specs/mountWrapper";
import { IStoreState, PluginNames, UnprocessableEntityError, UpgradeError } from "shared/types";
import { getType } from "typesafe-actions";
import actions from ".";

let store: any;

beforeEach(() => {
  store = getStore({
    apps: {
      ...initialState.apps,
      isFetching: false,
      items: [],
    },
    config: {
      ...initialState.config,
      namespace: "kubeapps-ns",
    },
  } as Partial<IStoreState>);
});

describe("fetches installed packages", () => {
  const validInstalledPackageSummary = new InstalledPackageSummary({
    installedPackageRef: new InstalledPackageReference({
      context: new Context({ cluster: "second-cluster", namespace: "my-ns" }),
      identifier: "some-name",
    }),
    iconUrl: "",
    name: "foo",
    pkgDisplayName: "foo",
    shortDescription: "some description",
  });
  let requestInstalledPackageListMock: jest.Mock;
  const installedPackageSummaries: InstalledPackageSummary[] = [validInstalledPackageSummary];
  beforeEach(() => {
    requestInstalledPackageListMock = jest.fn(
      () =>
        ({
          installedPackageSummaries,
        } as GetInstalledPackageSummariesResponse),
    );
    InstalledPackage.GetInstalledPackageSummaries = requestInstalledPackageListMock;
  });
  afterEach(() => {
    jest.restoreAllMocks();
  });
  it("fetches applications", async () => {
    const expectedActions = [
      {
        type: getType(actions.installedpackages.requestInstalledPackageList),
        payload: undefined,
        meta: undefined,
        error: undefined,
      },
      {
        type: getType(actions.installedpackages.receiveInstalledPackageList),
        payload: [validInstalledPackageSummary],
        meta: undefined,
        error: undefined,
      },
    ];
    await store.dispatch(
      actions.installedpackages.fetchInstalledPackages("second-cluster", "default"),
    );
    expect(store.getActions()).toEqual(expectedActions);
    expect(requestInstalledPackageListMock.mock.calls[0]).toEqual(["second-cluster", "default"]);
  });
});

describe("delete applications", () => {
  const deleteInstalledPackageOrig = InstalledPackage.DeleteInstalledPackage;
  let deleteInstalledPackage: jest.Mock;
  beforeEach(() => {
    deleteInstalledPackage = jest.fn(() => []);
    InstalledPackage.DeleteInstalledPackage = deleteInstalledPackage;
  });
  afterEach(() => {
    InstalledPackage.DeleteInstalledPackage = deleteInstalledPackageOrig;
  });
  it("delete an application", async () => {
    await store.dispatch(
      actions.installedpackages.deleteInstalledPackage({
        context: { cluster: "default-c", namespace: "default-ns" },
        identifier: "foo",
        plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
      } as InstalledPackageReference),
    );
    const expectedActions = [
      { type: getType(actions.installedpackages.requestDeleteInstalledPackage) },
      { type: getType(actions.installedpackages.receiveDeleteInstalledPackage) },
    ];
    expect(store.getActions()).toEqual(expectedActions);
    expect(deleteInstalledPackage.mock.calls[0]).toEqual([
      {
        context: { cluster: "default-c", namespace: "default-ns" },
        identifier: "foo",
        plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
      } as InstalledPackageReference,
    ]);
  });
  it("delete and throw an error", async () => {
    const error = new Error("something went wrong!");
    const expectedActions = [
      { type: getType(actions.installedpackages.requestDeleteInstalledPackage) },
      { type: getType(actions.installedpackages.errorInstalledPackage), payload: error },
    ];
    deleteInstalledPackage.mockImplementation(() => {
      throw error;
    });
    expect(
      await store.dispatch(
        actions.installedpackages.deleteInstalledPackage({
          context: { cluster: "default-c", namespace: "default-ns" },
          identifier: "foo",
          plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
        } as InstalledPackageReference),
      ),
    ).toBe(false);
    expect(store.getActions()).toEqual(expectedActions);
  });
});

describe("deploy package", () => {
  beforeEach(() => {
    InstalledPackage.CreateInstalledPackage = jest.fn();
  });

  it("returns true if namespace is correct and deployment is successful", async () => {
    const res = await store.dispatch(
      actions.installedpackages.installPackage(
        "target-cluster",
        "target-namespace",
        {
          name: "my-version",
          availablePackageRef: {
            identifier: "testrepo/foo",
          },
          version: {
            pkgVersion: "1.2.3",
            appVersion: "3.2.1",
          },
        } as AvailablePackageDetail,
        "my-release",
      ),
    );
    expect(res).toBe(true);
    expect(InstalledPackage.CreateInstalledPackage).toHaveBeenCalledWith(
      { cluster: "target-cluster", namespace: "target-namespace" },
      "my-release",
      { identifier: "testrepo/foo" },
      { version: "1.2.3" },
      undefined,
      undefined,
    );
    const expectedActions = [
      { type: getType(actions.installedpackages.requestInstallPackage) },
      { type: getType(actions.installedpackages.receiveInstallPackage) },
    ];
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("returns true if namespace is correct and deployment is successful with custom service account", async () => {
    const res = await store.dispatch(
      actions.installedpackages.installPackage(
        "target-cluster",
        "target-namespace",
        {
          name: "my-version",
          availablePackageRef: {
            identifier: "testrepo/foo",
          },
          version: {
            pkgVersion: "1.2.3",
            appVersion: "3.2.1",
          },
        } as AvailablePackageDetail,
        "my-release",
        undefined,
        undefined,
        { serviceAccountName: "my-sa" } as ReconciliationOptions,
      ),
    );
    expect(res).toBe(true);
    expect(InstalledPackage.CreateInstalledPackage).toHaveBeenCalledWith(
      { cluster: "target-cluster", namespace: "target-namespace" },
      "my-release",
      { identifier: "testrepo/foo" },
      { version: "1.2.3" },
      undefined,
      { serviceAccountName: "my-sa" } as ReconciliationOptions,
    );
    const expectedActions = [
      { type: getType(actions.installedpackages.requestInstallPackage) },
      { type: getType(actions.installedpackages.receiveInstallPackage) },
    ];
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("returns false and dispatches UnprocessableEntity if the given values don't satisfy the schema", async () => {
    const res = await store.dispatch(
      actions.installedpackages.installPackage(
        "target-cluster",
        "default",
        { name: "my-version" } as AvailablePackageDetail,
        "my-release",
        "foo: 1",
        {
          properties: { foo: { type: "string" } },
        } as any,
      ),
    );
    expect(res).toBe(false);
    const expectedActions = [
      { type: getType(actions.installedpackages.requestInstallPackage) },
      {
        type: getType(actions.installedpackages.errorInstalledPackage),
        payload: new UnprocessableEntityError(
          "The given values don't match the required format. The following errors were found:\n  - /foo: must be string",
        ),
      },
    ];
    expect(store.getActions()).toEqual(expectedActions);
  });
});

describe("updateInstalledPackage", () => {
  const updateInstalledPackageAction = actions.installedpackages.updateInstalledPackage(
    {
      context: { cluster: "default-c", namespace: "default-ns" },
      identifier: "my-release",
      plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
    } as InstalledPackageReference,
    { version: { appVersion: "4.5.6", pkgVersion: "1.2.3" } } as AvailablePackageDetail,
    "new-values",
  );

  it("calls updateInstalledPackage and returns true if no error", async () => {
    InstalledPackage.UpdateInstalledPackage = jest.fn().mockImplementationOnce(() => true);
    const res = await store.dispatch(updateInstalledPackageAction);
    expect(res).toBe(true);

    const expectedActions = [
      { type: getType(actions.installedpackages.requestUpdateInstalledPackage) },
      { type: getType(actions.installedpackages.receiveUpdateInstalledPackage) },
    ];
    expect(store.getActions()).toEqual(expectedActions);
    expect(InstalledPackage.UpdateInstalledPackage).toHaveBeenCalledWith(
      {
        context: { cluster: "default-c", namespace: "default-ns" },
        identifier: "my-release",
        plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
      } as InstalledPackageReference,
      { version: "1.2.3" } as VersionReference,
      "new-values",
    );
  });

  it("dispatches UpgradeError if error", async () => {
    InstalledPackage.UpdateInstalledPackage = jest.fn().mockImplementationOnce(() => {
      throw new UpgradeError("Boom!");
    });

    const expectedActions = [
      { type: getType(actions.installedpackages.requestUpdateInstalledPackage) },
      {
        type: getType(actions.installedpackages.errorInstalledPackage),
        payload: new UpgradeError("Boom!"),
      },
    ];

    await store.dispatch(updateInstalledPackageAction);
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("returns false and dispatches UnprocessableEntity if the given values don't satisfy the schema", async () => {
    const res = await store.dispatch(
      actions.installedpackages.updateInstalledPackage(
        {
          context: { cluster: "default-c", namespace: "default-ns" },
          identifier: "my-release",
          plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
        } as InstalledPackageReference,
        { version: { appVersion: "4.5.6", pkgVersion: "1.2.3" } } as AvailablePackageDetail,
        "foo: 1",
        {
          properties: { foo: { type: "string" } },
        } as any,
      ),
    );

    expect(res).toBe(false);
    const expectedActions = [
      { type: getType(actions.installedpackages.requestUpdateInstalledPackage) },
      {
        type: getType(actions.installedpackages.errorInstalledPackage),
        payload: new UnprocessableEntityError(
          "The given values don't match the required format. The following errors were found:\n  - /foo: must be string",
        ),
      },
    ];
    expect(store.getActions()).toEqual(expectedActions);
  });
});

describe("rollbackInstalledPackage", () => {
  const rollbackInstalledPackageAction = actions.installedpackages.rollbackInstalledPackage(
    {
      context: { cluster: "default-c", namespace: "default-ns" },
      identifier: "my-release",
      plugin: { name: PluginNames.PACKAGES_HELM, version: "0.0.1" } as Plugin,
    } as InstalledPackageReference,
    1,
  );

  it("success and re-request apps info", async () => {
    const installedPackageDetail = new InstalledPackageDetail({
      availablePackageRef: {
        context: { cluster: "default", namespace: "my-ns" },
        identifier: "test",
        plugin: { name: PluginNames.PACKAGES_HELM, version: "0.0.1" } as Plugin,
      },
      currentVersion: { appVersion: "4.5.6", pkgVersion: "1.2.3" },
    });

    const availablePackageDetail = new AvailablePackageDetail({ name: "test" });

    InstalledPackage.RollbackInstalledPackage = jest.fn().mockImplementationOnce(() => true);
    InstalledPackage.GetInstalledPackageDetail = jest.fn().mockReturnValue({
      installedPackageDetail: installedPackageDetail,
    });
    const res = await store.dispatch(rollbackInstalledPackageAction);
    expect(res).toBe(true);

    const selectCMD = actions.installedpackages.selectInstalledPackage(
      installedPackageDetail,
      availablePackageDetail,
    );
    const res2 = await store.dispatch(selectCMD);
    expect(res2).not.toBeNull();

    const expectedActions = [
      { type: getType(actions.installedpackages.requestRollbackInstalledPackage) },
      { type: getType(actions.installedpackages.receiveRollbackInstalledPackage) },
      { type: getType(actions.installedpackages.requestInstalledPackage) },
      {
        type: getType(actions.installedpackages.selectInstalledPackage),
        payload: { pkg: installedPackageDetail, details: availablePackageDetail },
      },
    ];

    expect(store.getActions()).toEqual(expectedActions);
    expect(InstalledPackage.RollbackInstalledPackage).toHaveBeenCalledWith(
      {
        context: { cluster: "default-c", namespace: "default-ns" },
        identifier: "my-release",
        plugin: { name: PluginNames.PACKAGES_HELM, version: "0.0.1" },
      },
      1,
    );
  });

  it("dispatches an error if the package is not from one of the supported plugins", async () => {
    const expectedActions = [
      {
        type: getType(actions.installedpackages.errorInstalledPackage),
        payload: new UpgradeError(
          "This package cannot be rolled back; this operation is only available for Helm packages",
        ),
      },
    ];

    const rollbackInstalledPackageBadAction = actions.installedpackages.rollbackInstalledPackage(
      {
        context: { cluster: "default-c", namespace: "default-ns" },
        identifier: "my-release",
        plugin: { name: "bad-plugin", version: "0.0.1" } as Plugin,
      } as InstalledPackageReference,
      1,
    );
    await store.dispatch(rollbackInstalledPackageBadAction);
    expect(store.getActions()).toEqual(expectedActions);
  });
});
