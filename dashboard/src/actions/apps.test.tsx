import {
  AvailablePackageDetail,
  InstalledPackageDetail,
  InstalledPackageSummary,
  GetInstalledPackageSummariesResponse,
  InstalledPackageReference,
  VersionReference,
  ReconciliationOptions,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";
import { App } from "shared/App";
import { IAppState, UnprocessableEntity, UpgradeError } from "shared/types";
import { PluginNames } from "shared/utils";
import { getType } from "typesafe-actions";
import actions from ".";

const mockStore = configureMockStore([thunk]);

let store: any;

beforeEach(() => {
  const state: IAppState = {
    isFetching: false,
    items: [],
  };
  store = mockStore({
    apps: {
      state,
    },
    config: {
      namespace: "kubeapps-ns",
    },
  });
});

describe("fetches applications", () => {
  const validInstalledPackageSummary: InstalledPackageSummary = {
    installedPackageRef: {
      context: { cluster: "second-cluster", namespace: "my-ns" },
      identifier: "some-name",
    },
    iconUrl: "",
    name: "foo",
    pkgDisplayName: "foo",
    shortDescription: "some description",
  };
  let listAppsMock: jest.Mock;
  const installedPackageSummaries: InstalledPackageSummary[] = [validInstalledPackageSummary];
  beforeEach(() => {
    listAppsMock = jest.fn(
      () =>
        ({
          installedPackageSummaries,
        } as GetInstalledPackageSummariesResponse),
    );
    App.GetInstalledPackageSummaries = listAppsMock;
  });
  afterEach(() => {
    jest.restoreAllMocks();
  });
  it("fetches applications", async () => {
    const expectedActions = [
      {
        type: getType(actions.apps.listApps),
        payload: undefined,
        meta: undefined,
        error: undefined,
      },
      {
        type: getType(actions.apps.receiveAppList),
        payload: [validInstalledPackageSummary],
        meta: undefined,
        error: undefined,
      },
    ];
    await store.dispatch(actions.apps.fetchApps("second-cluster", "default"));
    expect(store.getActions()).toEqual(expectedActions);
    expect(listAppsMock.mock.calls[0]).toEqual(["second-cluster", "default"]);
  });
});

describe("delete applications", () => {
  const deleteInstalledPackageOrig = App.DeleteInstalledPackage;
  let deleteInstalledPackage: jest.Mock;
  beforeEach(() => {
    deleteInstalledPackage = jest.fn(() => []);
    App.DeleteInstalledPackage = deleteInstalledPackage;
  });
  afterEach(() => {
    App.DeleteInstalledPackage = deleteInstalledPackageOrig;
  });
  it("delete an application", async () => {
    await store.dispatch(
      actions.apps.deleteInstalledPackage({
        context: { cluster: "default-c", namespace: "default-ns" },
        identifier: "foo",
        plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
      } as InstalledPackageReference),
    );
    const expectedActions = [
      { type: getType(actions.apps.requestDeleteInstalledPackage) },
      { type: getType(actions.apps.receiveDeleteInstalledPackage) },
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
      { type: getType(actions.apps.requestDeleteInstalledPackage) },
      { type: getType(actions.apps.errorApp), payload: error },
    ];
    deleteInstalledPackage.mockImplementation(() => {
      throw error;
    });
    expect(
      await store.dispatch(
        actions.apps.deleteInstalledPackage({
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
    App.CreateInstalledPackage = jest.fn();
  });

  it("returns true if namespace is correct and deployment is successful", async () => {
    const res = await store.dispatch(
      actions.apps.installPackage(
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
    expect(App.CreateInstalledPackage).toHaveBeenCalledWith(
      { cluster: "target-cluster", namespace: "target-namespace" },
      "my-release",
      { identifier: "testrepo/foo" },
      { version: "1.2.3" },
      undefined,
      undefined,
    );
    const expectedActions = [
      { type: getType(actions.apps.requestInstallPackage) },
      { type: getType(actions.apps.receiveInstallPackage) },
    ];
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("returns true if namespace is correct and deployment is successful with custom service account", async () => {
    const res = await store.dispatch(
      actions.apps.installPackage(
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
    expect(App.CreateInstalledPackage).toHaveBeenCalledWith(
      { cluster: "target-cluster", namespace: "target-namespace" },
      "my-release",
      { identifier: "testrepo/foo" },
      { version: "1.2.3" },
      undefined,
      { serviceAccountName: "my-sa" } as ReconciliationOptions,
    );
    const expectedActions = [
      { type: getType(actions.apps.requestInstallPackage) },
      { type: getType(actions.apps.receiveInstallPackage) },
    ];
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("returns false and dispatches UnprocessableEntity if the given values don't satisfy the schema", async () => {
    const res = await store.dispatch(
      actions.apps.installPackage(
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
      { type: getType(actions.apps.requestInstallPackage) },
      {
        type: getType(actions.apps.errorApp),
        payload: new UnprocessableEntity(
          "The given values don't match the required format. The following errors were found:\n  - /foo: must be string",
        ),
      },
    ];
    expect(store.getActions()).toEqual(expectedActions);
  });
});

describe("updateInstalledPackage", () => {
  const updateInstalledPackageAction = actions.apps.updateInstalledPackage(
    {
      context: { cluster: "default-c", namespace: "default-ns" },
      identifier: "my-release",
      plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
    } as InstalledPackageReference,
    { version: { appVersion: "4.5.6", pkgVersion: "1.2.3" } } as AvailablePackageDetail,
    "new-values",
  );

  it("calls updateInstalledPackage and returns true if no error", async () => {
    App.UpdateInstalledPackage = jest.fn().mockImplementationOnce(() => true);
    const res = await store.dispatch(updateInstalledPackageAction);
    expect(res).toBe(true);

    const expectedActions = [
      { type: getType(actions.apps.requestUpdateInstalledPackage) },
      { type: getType(actions.apps.receiveUpdateInstalledPackage) },
    ];
    expect(store.getActions()).toEqual(expectedActions);
    expect(App.UpdateInstalledPackage).toHaveBeenCalledWith(
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
    App.UpdateInstalledPackage = jest.fn().mockImplementationOnce(() => {
      throw new UpgradeError("Boom!");
    });

    const expectedActions = [
      { type: getType(actions.apps.requestUpdateInstalledPackage) },
      {
        type: getType(actions.apps.errorApp),
        payload: new UpgradeError("Boom!"),
      },
    ];

    await store.dispatch(updateInstalledPackageAction);
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("returns false and dispatches UnprocessableEntity if the given values don't satisfy the schema", async () => {
    const res = await store.dispatch(
      actions.apps.updateInstalledPackage(
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
      { type: getType(actions.apps.requestUpdateInstalledPackage) },
      {
        type: getType(actions.apps.errorApp),
        payload: new UnprocessableEntity(
          "The given values don't match the required format. The following errors were found:\n  - /foo: must be string",
        ),
      },
    ];
    expect(store.getActions()).toEqual(expectedActions);
  });
});

describe("rollbackInstalledPackage", () => {
  const rollbackInstalledPackageAction = actions.apps.rollbackInstalledPackage(
    {
      context: { cluster: "default-c", namespace: "default-ns" },
      identifier: "my-release",
      plugin: { name: PluginNames.PACKAGES_HELM, version: "0.0.1" } as Plugin,
    } as InstalledPackageReference,
    1,
  );

  it("success and re-request apps info", async () => {
    const installedPackageDetail = {
      availablePackageRef: {
        context: { cluster: "default", namespace: "my-ns" },
        identifier: "test",
        plugin: { name: PluginNames.PACKAGES_HELM, version: "0.0.1" } as Plugin,
      },
      currentVersion: { appVersion: "4.5.6", pkgVersion: "1.2.3" },
    } as InstalledPackageDetail;

    const availablePackageDetail = { name: "test" } as AvailablePackageDetail;

    App.RollbackInstalledPackage = jest.fn().mockImplementationOnce(() => true);
    App.GetInstalledPackageResourceRefs = jest.fn().mockReturnValue({ resourceRefs: [] });
    App.GetInstalledPackageDetail = jest.fn().mockReturnValue({
      installedPackageDetail: installedPackageDetail,
    });
    const res = await store.dispatch(rollbackInstalledPackageAction);
    expect(res).toBe(true);

    const selectCMD = actions.apps.selectApp(
      installedPackageDetail as any,
      [],
      availablePackageDetail,
    );
    const res2 = await store.dispatch(selectCMD);
    expect(res2).not.toBeNull();

    const expectedActions = [
      { type: getType(actions.apps.requestRollbackInstalledPackage) },
      { type: getType(actions.apps.receiveRollbackInstalledPackage) },
      { type: getType(actions.apps.requestApps) },
      {
        type: getType(actions.apps.selectApp),
        payload: { app: installedPackageDetail, resourceRefs: [], details: availablePackageDetail },
      },
    ];

    expect(store.getActions()).toEqual(expectedActions);
    expect(App.RollbackInstalledPackage).toHaveBeenCalledWith(
      {
        context: { cluster: "default-c", namespace: "default-ns" },
        identifier: "my-release",
        plugin: { name: PluginNames.PACKAGES_HELM, version: "0.0.1" },
      },
      1,
    );
    expect(App.GetInstalledPackageResourceRefs).toHaveBeenCalledWith({
      context: { cluster: "default-c", namespace: "default-ns" },
      identifier: "my-release",
      plugin: { name: PluginNames.PACKAGES_HELM, version: "0.0.1" },
    });
  });

  it("dispatches an error if the package is not from one of the supported plugins", async () => {
    const expectedActions = [
      {
        type: getType(actions.apps.errorApp),
        payload: new UpgradeError(
          "This package cannot be rolled back; this operation is only available for Helm packages",
        ),
      },
    ];

    const rollbackInstalledPackageBadAction = actions.apps.rollbackInstalledPackage(
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
