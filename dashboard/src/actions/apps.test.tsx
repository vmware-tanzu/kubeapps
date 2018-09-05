import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";
import { getType } from "typesafe-actions";

import actions from ".";
import { App } from "../shared/App";
import { IAppState } from "../shared/types";

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
