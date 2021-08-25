import { IAppState } from "shared/types";
import { getType } from "typesafe-actions";
import actions from "../actions";
import appsReducer from "./apps";

describe("appsReducer", () => {
  let initialState: IAppState;

  const actionTypes = {
    listApps: getType(actions.apps.listApps),
    requestApps: getType(actions.apps.requestApps),
  };

  beforeEach(() => {
    initialState = {
      isFetching: false,
      items: [],
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
        type: actionTypes.listApps as any,
      });
      expect(state).toEqual({ ...initialState, isFetching: true });
      state = appsReducer(state, {
        type: actionTypes.listApps as any,
      });
      expect(state).toEqual({ ...initialState, isFetching: true });
    });
  });
});
