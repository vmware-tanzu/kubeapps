import { AvailablePackageDetail } from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";
import { App } from "shared/App";
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
    jest.clearAllMocks();
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
        {
          name: "my-version",
        } as AvailablePackageDetail,
        "my-release",
      ),
    );
    expect(res).toBe(true);
    expect(App.create).toHaveBeenCalledWith(
      "target-cluster",
      "target-namespace",
      "my-release",
      {
        name: "my-version",
      } as AvailablePackageDetail,
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
  const provisionCMD = actions.apps.rollbackApp("default-c", "default-ns", "my-release", 1);

  it("success and re-request apps info", async () => {
    App.rollback = jest.fn().mockImplementationOnce(() => true);
    App.getRelease = jest.fn().mockImplementationOnce(() => true);
    const res = await store.dispatch(provisionCMD);
    expect(res).toBe(true);

    const expectedActions = [
      { type: getType(actions.apps.requestRollbackApp) },
      { type: getType(actions.apps.receiveRollbackApp) },
      // TODO(agamez): check if we really should dispatch this requestApps action after a rollback
      // { type: getType(actions.apps.requestApps) },
      // { type: getType(actions.apps.selectApp), payload: true },
    ];
    expect(store.getActions()).toEqual(expectedActions);
    expect(App.rollback).toHaveBeenCalledWith("default-c", "default-ns", "my-release", 1);
    // expect(App.getRelease).toHaveBeenCalledWith("default-c", "default-ns", "my-release");
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
