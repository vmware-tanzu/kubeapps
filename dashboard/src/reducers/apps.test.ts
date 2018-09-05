import { getType } from "typesafe-actions";
import actions from "../actions";
import { IAppState } from "../shared/types";
import appsReducer from "./apps";

describe("authReducer", () => {
  let initialState: IAppState;

  const actionTypes = {
    listApps: getType(actions.apps.listApps),
    receiveAppList: getType(actions.apps.receiveAppList),
    requestApps: getType(actions.apps.requestApps),
  };

  beforeEach(() => {
    initialState = {
      isFetching: false,
      items: [],
      listingAll: false,
    };
  });

  describe("reducer actions", () => {
    it("sets isFetching when requesting an app", () => {
      [true, false].forEach(e => {
        expect(
          appsReducer(undefined, {
            type: actionTypes.requestApps as any,
          }),
        ).toEqual({ ...initialState, isFetching: true });
      });
    });

    it("toggles the listAll state", () => {
      let state = appsReducer(undefined, {
        listingAll: true,
        type: actionTypes.listApps as any,
      });
      expect(state).toEqual({ ...initialState, isFetching: true, listingAll: true });
      state = appsReducer(state, {
        listingAll: false,
        type: actionTypes.listApps as any,
      });
      expect(state).toEqual({ ...initialState, isFetching: true, listingAll: false });
    });
  });
});
