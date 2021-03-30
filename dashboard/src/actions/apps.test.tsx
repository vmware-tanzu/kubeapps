import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";
import { getType } from "typesafe-actions";

import actions from ".";
import { App } from "../shared/App";
import Chart from "../shared/Chart";
import { IAppState, UnprocessableEntity } from "../shared/types";

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
    App.listApps = listAppsMock;
  });
  afterEach(() => {
    jest.clearAllMocks();
  });
  it("fetches applications", async () => {
    const expectedActions = [
      { type: getType(actions.apps.listApps) },
      { type: getType(actions.apps.receiveAppList), payload: [] },
    ];
    await store.dispatch(actions.apps.fetchAppsWithUpdateInfo("default-cluster", "default"));
    expect(store.getActions()).toEqual(expectedActions);
    expect(listAppsMock.mock.calls[0]).toEqual(["default-cluster", "default"]);
  });
  it("fetches applications, ignore when no data", async () => {
    App.listApps = jest.fn();
    const expectedActions = [
      { type: getType(actions.apps.listApps) },
      { type: getType(actions.apps.receiveAppList), payload: undefined },
    ];
    await store.dispatch(actions.apps.fetchAppsWithUpdateInfo("default-cluster", "default"));
    App.listApps = listAppsMock;
    expect(store.getActions()).toEqual(expectedActions);
    expect(listAppsMock.mock.calls[0]).toBeUndefined();
  });

  describe("fetches chart updates", () => {
    it("gets a chart latest version", async () => {
      const appsResponse = [
        {
          releaseName: "foobar",
          namespace: "ns-1",
          chartMetadata: { name: "foo", version: "1.0.0", appVersion: "0.1.0" },
        },
      ];
      const chartUpdatesResponse = [
        {
          attributes: { repo: { name: "bar" } },
          relationships: {
            latestChartVersion: { data: { app_version: "1.0.0", version: "1.1.0" } },
          },
        },
      ];
      Chart.listWithFilters = jest.fn().mockReturnValue(chartUpdatesResponse);
      App.listApps = jest.fn().mockReturnValue(appsResponse);
      const expectedActions = [
        { type: getType(actions.apps.listApps) },
        { type: getType(actions.apps.receiveAppList), payload: appsResponse },
        { type: getType(actions.apps.requestAppUpdateInfo) },
        {
          type: getType(actions.apps.receiveAppUpdateInfo),
          payload: {
            releaseName: "foobar",
            updateInfo: {
              upToDate: false,
              appLatestVersion: "1.0.0",
              chartLatestVersion: "1.1.0",
              repository: { name: "bar" },
            },
          },
        },
      ];
      await store.dispatch(actions.apps.fetchAppsWithUpdateInfo("default-c", "default-ns"));
      // It should use the app namespace
      expect(Chart.listWithFilters).toHaveBeenCalledWith(
        "default-c",
        "ns-1",
        "foo",
        "1.0.0",
        "0.1.0",
      );
      expect(store.getActions()).toEqual(expectedActions);
    });

    it("set up upToDate=true if the application is up to date", async () => {
      const appsResponse = [
        {
          releaseName: "foobar",
          chartMetadata: { name: "foo", version: "1.0.0", appVersion: "0.1.0" },
        },
      ];
      const chartUpdatesResponse = [
        {
          attributes: { repo: { name: "bar" } },
          relationships: {
            latestChartVersion: { data: { app_version: "0.1.0", version: "1.0.0" } },
          },
        },
      ];
      Chart.listWithFilters = jest.fn().mockReturnValue(chartUpdatesResponse);
      App.listApps = jest.fn().mockReturnValue(appsResponse);
      const expectedActions = [
        { type: getType(actions.apps.listApps) },
        { type: getType(actions.apps.receiveAppList), payload: appsResponse },
        { type: getType(actions.apps.requestAppUpdateInfo) },
        {
          type: getType(actions.apps.receiveAppUpdateInfo),
          payload: {
            releaseName: "foobar",
            updateInfo: {
              upToDate: true,
              appLatestVersion: "0.1.0",
              chartLatestVersion: "1.0.0",
              repository: { name: "bar" },
            },
          },
        },
      ];
      await store.dispatch(actions.apps.fetchAppsWithUpdateInfo("default-c", "default-ns"));
      expect(store.getActions()).toEqual(expectedActions);
    });

    it("set an error if the application version is not semver compatible", async () => {
      const appsResponse = [
        {
          releaseName: "foobar",
          chartMetadata: { name: "foo", version: "1.0", appVersion: "0.1.0" },
        },
      ];
      const chartUpdatesResponse = [
        {
          attributes: { repo: { name: "bar" } },
          relationships: { latestChartVersion: { data: { version: "1.0" } } },
        },
      ];
      Chart.listWithFilters = jest.fn().mockReturnValue(chartUpdatesResponse);
      App.listApps = jest.fn().mockReturnValue(appsResponse);
      const expectedActions = [
        { type: getType(actions.apps.listApps) },
        { type: getType(actions.apps.receiveAppList), payload: appsResponse },
        { type: getType(actions.apps.requestAppUpdateInfo) },
        {
          type: getType(actions.apps.receiveAppUpdateInfo),
          payload: {
            releaseName: "foobar",
            updateInfo: {
              error: new Error("Invalid Version: 1.0"),
              upToDate: false,
              chartLatestVersion: "",
              appLatestVersion: "",
              repository: { name: "", url: "", namespace: "" },
            },
          },
        },
      ];
      await store.dispatch(actions.apps.fetchAppsWithUpdateInfo("default-c", "default-ns"));
      expect(store.getActions()).toEqual(expectedActions);
    });
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
    await store.dispatch(actions.apps.deleteApp("default-c", "default-ns", "foo", false));
    const expectedActions = [
      { type: getType(actions.apps.requestDeleteApp) },
      { type: getType(actions.apps.receiveDeleteApp) },
    ];
    expect(store.getActions()).toEqual(expectedActions);
    expect(deleteAppMock.mock.calls[0]).toEqual(["default-c", "default-ns", "foo", false]);
  });
  it("delete and purge an application", async () => {
    await store.dispatch(actions.apps.deleteApp("default-c", "default-ns", "foo", true));
    const expectedActions = [
      { type: getType(actions.apps.requestDeleteApp) },
      { type: getType(actions.apps.receiveDeleteApp) },
    ];
    expect(store.getActions()).toEqual(expectedActions);
    expect(deleteAppMock.mock.calls[0]).toEqual(["default-c", "default-ns", "foo", true]);
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
      await store.dispatch(actions.apps.deleteApp("default-c", "default-ns", "foo", true)),
    ).toBe(false);
    expect(store.getActions()).toEqual(expectedActions);
  });
});

describe("deploy chart", () => {
  beforeEach(() => {
    App.create = jest.fn();
  });

  it("returns true if namespace is correct and deployment is successful", async () => {
    const res = await store.dispatch(
      actions.apps.deployChart(
        "target-cluster",
        "target-namespace",
        "my-version" as any,
        "chart-namespace",
        "my-release",
      ),
    );
    expect(res).toBe(true);
    expect(App.create).toHaveBeenCalledWith(
      "target-cluster",
      "target-namespace",
      "my-release",
      "chart-namespace",
      "my-version",
      undefined,
    );
    const expectedActions = [
      { type: getType(actions.apps.requestDeployApp) },
      { type: getType(actions.apps.receiveDeployApp) },
    ];
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("returns false and dispatches UnprocessableEntity if the given values don't satisfy the schema", async () => {
    const res = await store.dispatch(
      actions.apps.deployChart(
        "target-cluster",
        "default",
        "my-version" as any,
        "chart-namespace",
        "my-release",
        "foo: 1",
        {
          properties: { foo: { type: "string" } },
        },
      ),
    );
    expect(res).toBe(false);
    const expectedActions = [
      { type: getType(actions.apps.requestDeployApp) },
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
    "default-c",
    "kubeapps-ns",
    "my-version" as any,
    "kubeapps",
    "my-release",
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
      "default-c",
      "kubeapps-ns",
      "my-release",
      "kubeapps",
      "my-version" as any,
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
        "default-c",
        "kubeapps-ns",
        "my-version" as any,
        "default",
        "my-release",
        "foo: 1",
        {
          properties: { foo: { type: "string" } },
        },
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
  const provisionCMD = actions.apps.rollbackApp("default-c", "default-ns", "my-release", 1);

  it("success and re-request apps info", async () => {
    App.rollback = jest.fn().mockImplementationOnce(() => true);
    App.getRelease = jest.fn().mockImplementationOnce(() => true);
    const res = await store.dispatch(provisionCMD);
    expect(res).toBe(true);

    const expectedActions = [
      { type: getType(actions.apps.requestRollbackApp) },
      { type: getType(actions.apps.receiveRollbackApp) },
      { type: getType(actions.apps.requestApps) },
      { type: getType(actions.apps.selectApp), payload: true },
    ];
    expect(store.getActions()).toEqual(expectedActions);
    expect(App.rollback).toHaveBeenCalledWith("default-c", "default-ns", "my-release", 1);
    expect(App.getRelease).toHaveBeenCalledWith("default-c", "default-ns", "my-release");
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
