import {
  AvailablePackageDetail,
  InstalledPackageDetail,
  InstalledPackageReference,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";
import { App } from "shared/App";
import Chart from "shared/Chart";
import { IAppState, UnprocessableEntity } from "shared/types";
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
  let listAppsMock: jest.Mock;
  beforeEach(() => {
    listAppsMock = jest.fn(() => []);
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
        payload: undefined,
        meta: undefined,
        error: undefined,
      },
    ];
    await store.dispatch(actions.apps.fetchApps("default-cluster", "default"));
    expect(store.getActions()).toEqual(expectedActions);
    expect(listAppsMock.mock.calls[0]).toEqual(["default-cluster", "default"]);
  });
  it("fetches applications, ignore when no data", async () => {
    App.GetInstalledPackageSummaries = jest.fn();
    const expectedActions = [
      {
        type: getType(actions.apps.listApps),
        payload: undefined,
        meta: undefined,
        error: undefined,
      },
      {
        type: getType(actions.apps.receiveAppList),
        payload: undefined,
        meta: undefined,
        error: undefined,
      },
    ];
    await store.dispatch(actions.apps.fetchApps("default-cluster", "default"));
    App.GetInstalledPackageSummaries = listAppsMock;
    expect(store.getActions()).toEqual(expectedActions);
    expect(listAppsMock.mock.calls[0]).toBeUndefined();
  });
});

describe("delete applications", () => {
  const deleteAppOrig = App.delete;
  let deleteAppMock: jest.Mock;
  beforeEach(() => {
    deleteAppMock = jest.fn(() => []);
    App.delete = deleteAppMock;
  });
  afterEach(() => {
    App.delete = deleteAppOrig;
  });
  it("delete an application", async () => {
    await store.dispatch(
      actions.apps.deleteApp(
        {
          context: { cluster: "default-c", namespace: "default-ns" },
          identifier: "foo",
          plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
        } as InstalledPackageReference,
        false,
      ),
    );
    const expectedActions = [
      { type: getType(actions.apps.requestDeleteApp) },
      { type: getType(actions.apps.receiveDeleteApp) },
    ];
    expect(store.getActions()).toEqual(expectedActions);
    expect(deleteAppMock.mock.calls[0]).toEqual([
      {
        context: { cluster: "default-c", namespace: "default-ns" },
        identifier: "foo",
        plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
      } as InstalledPackageReference,
      false,
    ]);
  });
  it("delete and purge an application", async () => {
    await store.dispatch(
      actions.apps.deleteApp(
        {
          context: { cluster: "default-c", namespace: "default-ns" },
          identifier: "foo",
          plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
        } as InstalledPackageReference,
        true,
      ),
    );
    const expectedActions = [
      { type: getType(actions.apps.requestDeleteApp) },
      { type: getType(actions.apps.receiveDeleteApp) },
    ];
    expect(store.getActions()).toEqual(expectedActions);
    expect(deleteAppMock.mock.calls[0]).toEqual([
      {
        context: { cluster: "default-c", namespace: "default-ns" },
        identifier: "foo",
        plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
      } as InstalledPackageReference,
      true,
    ]);
  });
  it("delete and throw an error", async () => {
    const error = new Error("something went wrong!");
    const expectedActions = [
      { type: getType(actions.apps.requestDeleteApp) },
      { type: getType(actions.apps.errorApp), payload: error },
    ];
    deleteAppMock.mockImplementation(() => {
      throw error;
    });
    expect(
      await store.dispatch(
        actions.apps.deleteApp(
          {
            context: { cluster: "default-c", namespace: "default-ns" },
            identifier: "foo",
            plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
          } as InstalledPackageReference,
          true,
        ),
      ),
    ).toBe(false);
    expect(store.getActions()).toEqual(expectedActions);
  });
});

describe("deploy chart", () => {
  beforeEach(() => {
    App.createInstalledPackage = jest.fn();
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
    expect(App.createInstalledPackage).toHaveBeenCalledWith(
      { cluster: "target-cluster", namespace: "target-namespace" },
      "my-release",
      { identifier: "testrepo/foo" },
      { version: "1.2.3" },
      undefined,
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

describe("upgradeApp", () => {
  const provisionCMD = actions.apps.upgradeApp(
    {
      context: { cluster: "default-c", namespace: "default-ns" },
      identifier: "my-release",
      plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
    } as InstalledPackageReference,

    {} as AvailablePackageDetail,
    "kubeapps",
  );

  it("calls ServiceBinding.delete and returns true if no error", async () => {
    App.upgrade = jest.fn().mockImplementationOnce(() => true);
    const res = await store.dispatch(provisionCMD);
    expect(res).toBe(true);

    const expectedActions = [
      { type: getType(actions.apps.requestUpgradeApp) },
      { type: getType(actions.apps.receiveUpgradeApp) },
    ];
    expect(store.getActions()).toEqual(expectedActions);
    expect(App.upgrade).toHaveBeenCalledWith(
      {
        context: { cluster: "default-c", namespace: "default-ns" },
        identifier: "my-release",
        plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
      } as InstalledPackageReference,
      "kubeapps",
      {} as AvailablePackageDetail,
      undefined,
    );
  });

  it("dispatches errorCatalog if error", async () => {
    App.upgrade = jest.fn().mockImplementationOnce(() => {
      throw new Error("Boom!");
    });

    const expectedActions = [
      { type: getType(actions.apps.requestUpgradeApp) },
      {
        type: getType(actions.apps.errorApp),
        payload: new Error("Boom!"),
      },
    ];

    await store.dispatch(provisionCMD);
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("returns false and dispatches UnprocessableEntity if the given values don't satisfy the schema", async () => {
    const res = await store.dispatch(
      actions.apps.upgradeApp(
        {
          context: { cluster: "default-c", namespace: "default-ns" },
          identifier: "my-release",
          plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
        } as InstalledPackageReference,
        {} as AvailablePackageDetail,
        "kubeapps",
        "foo: 1",
        {
          properties: { foo: { type: "string" } },
        } as any,
      ),
    );

    expect(res).toBe(false);
    const expectedActions = [
      { type: getType(actions.apps.requestUpgradeApp) },
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

describe("rollbackApp", () => {
  const provisionCMD = actions.apps.rollbackApp(
    {
      context: { cluster: "default-c", namespace: "default-ns" },
      identifier: "my-release",
      plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
    } as InstalledPackageReference,
    1,
  );

  it("success and re-request apps info", async () => {
    const installedPackageDetail = {
      availablePackageRef: {
        context: { cluster: "default", namespace: "my-ns" },
        identifier: "test",
        plugin: { name: "my.plugin", version: "0.0.1" } as Plugin,
      },
      currentVersion: { appVersion: "4.5.6", pkgVersion: "1.2.3" },
    } as InstalledPackageDetail;

    const availablePackageDetail = { name: "test" } as AvailablePackageDetail;

    App.rollback = jest.fn().mockImplementationOnce(() => true);
    App.getRelease = jest.fn().mockReturnValue({ manifest: {} });
    App.GetInstalledPackageDetail = jest.fn().mockReturnValue({
      installedPackageDetail: installedPackageDetail,
    });
    Chart.getAvailablePackageDetail = jest.fn().mockReturnValue({
      availablePackageDetail: availablePackageDetail,
    });
    const res = await store.dispatch(provisionCMD);
    expect(res).toBe(true);

    const selectCMD = actions.apps.selectApp(
      installedPackageDetail as any,
      {},
      availablePackageDetail,
    );
    const res2 = await store.dispatch(selectCMD);
    expect(res2).not.toBeNull();

    const expectedActions = [
      { type: getType(actions.apps.requestRollbackApp) },
      { type: getType(actions.apps.receiveRollbackApp) },
      { type: getType(actions.apps.requestApps) },
      {
        type: getType(actions.apps.selectApp),
        payload: { app: installedPackageDetail, manifest: {}, details: availablePackageDetail },
      },
    ];

    expect(store.getActions()).toEqual(expectedActions);
    expect(App.rollback).toHaveBeenCalledWith(
      {
        context: { cluster: "default-c", namespace: "default-ns" },
        identifier: "my-release",
        plugin: { name: "my.plugin", version: "0.0.1" },
      },
      1,
    );
    expect(App.getRelease).toHaveBeenCalledWith({
      context: { cluster: "default-c", namespace: "default-ns" },
      identifier: "my-release",
      plugin: { name: "my.plugin", version: "0.0.1" },
    });
  });

  it("dispatches an error", async () => {
    App.rollback = jest.fn().mockImplementationOnce(() => {
      throw new Error("Boom!");
    });

    const expectedActions = [
      { type: getType(actions.apps.requestRollbackApp) },
      {
        type: getType(actions.apps.errorApp),
        payload: new Error("Boom!"),
      },
    ];

    await store.dispatch(provisionCMD);
    expect(store.getActions()).toEqual(expectedActions);
  });
});
