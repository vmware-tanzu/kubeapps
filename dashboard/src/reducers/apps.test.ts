import { getType } from "typesafe-actions";
import actions from "../actions";
import { IAppState } from "../shared/types";
import appsReducer from "./apps";

describe("authReducer", () => {
  let initialState: IAppState;

  const actionTypes = {
    errorApps: getType(actions.apps.errorApps),
    errorDeleteApp: getType(actions.apps.errorDeleteApp),
    listApps: getType(actions.apps.listApps),
    receiveAppList: getType(actions.apps.receiveAppList),
    receiveApps: getType(actions.apps.receiveApps),
    requestApps: getType(actions.apps.requestApps),
    selectApp: getType(actions.apps.selectApp),
    toggleListAllAction: getType(actions.apps.toggleListAllAction),
  };

  beforeEach(() => {
    initialState = {
      isFetching: false,
      items: [],
      listAll: false,
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
        type: actionTypes.toggleListAllAction as any,
      });
      expect(state).toEqual({ ...initialState, listAll: true });
      state = appsReducer(state, {
        type: actionTypes.toggleListAllAction as any,
      });
      expect(state).toEqual({ ...initialState, listAll: false });
    });
  });
});
