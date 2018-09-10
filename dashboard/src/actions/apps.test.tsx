import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";
import { getType } from "typesafe-actions";

import actions from ".";
import { App } from "../shared/App";
import { definedNamespaces } from "../shared/Namespace";
import { IAppState, UnprocessableEntity } from "../shared/types";

const mockStore = configureMockStore([thunk]);

let store: any;

beforeEach(() => {
  const state: IAppState = {
    isFetching: false,
    items: [],
    listingAll: false,
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
  const listAppsOrig = App.listApps;
  let listAppsMock: jest.Mock;
  beforeEach(() => {
    listAppsMock = jest.fn(() => []);
    App.listApps = listAppsMock;
  });
  afterEach(() => {
    App.listApps = listAppsOrig;
  });
  it("fetches all applications", async () => {
    const expectedActions = [
      { type: getType(actions.apps.listApps), listingAll: true },
      { type: getType(actions.apps.receiveAppList), apps: [] },
    ];
    await store.dispatch(actions.apps.fetchApps("default", true));
    expect(store.getActions()).toEqual(expectedActions);
    expect(listAppsMock.mock.calls[0]).toEqual(["default", true]);
  });
  it("fetches default applications", () => {
    const expectedActions = [
      { type: getType(actions.apps.listApps), listingAll: false },
      { type: getType(actions.apps.receiveAppList), apps: [] },
    ];
    return store.dispatch(actions.apps.fetchApps("default", false)).then(() => {
      expect(store.getActions()).toEqual(expectedActions);
      expect(listAppsMock.mock.calls[0]).toEqual(["default", false]);
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
    await store.dispatch(actions.apps.deleteApp("foo", "default", false));
    expect(store.getActions()).toEqual([]);
    expect(deleteAppMock.mock.calls[0]).toEqual(["foo", "default", false]);
  });
  it("delete and purge an application", async () => {
    await store.dispatch(actions.apps.deleteApp("foo", "default", true));
    expect(store.getActions()).toEqual([]);
    expect(deleteAppMock.mock.calls[0]).toEqual(["foo", "default", true]);
  });
  it("delete and throw an error", async () => {
    const error = new Error("something went wrong!");
    const expectedActions = [{ type: getType(actions.apps.errorDeleteApp), err: error }];
    deleteAppMock.mockImplementation(() => {
      throw error;
    });
    expect(await store.dispatch(actions.apps.deleteApp("foo", "default", true))).toBe(false);
    expect(store.getActions()).toEqual(expectedActions);
  });
});

describe("deploy chart", () => {
  beforeEach(() => {
    App.create = jest.fn();
  });

  it("returns true if namespace is correct and deployment is successful", async () => {
    const res = await store.dispatch(
      actions.apps.deployChart("my-version" as any, "my-release", "default"),
    );
    expect(res).toBe(true);
    expect(App.create).toHaveBeenCalledWith(
      "my-release",
      "default",
      "kubeapps-ns",
      "my-version",
      undefined,
    );
    expect(store.getActions().length).toBe(0);
  });

  it("returns false and dispatches UnprocessableEntity if the namespace is _all", async () => {
    const res = await store.dispatch(
      actions.apps.deployChart("my-version" as any, "my-release", definedNamespaces.all),
    );
    expect(res).toBe(false);
    expect(store.getActions().length).toBe(1);
    expect(store.getActions()[0].type).toEqual(getType(actions.apps.errorApps));
    expect(store.getActions()[0].err.constructor).toBe(UnprocessableEntity);
  });
});
