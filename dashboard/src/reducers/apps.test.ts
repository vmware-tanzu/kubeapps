import { getType } from "typesafe-actions";
import actions from "../actions";
import { IAppState } from "../shared/types";
import appsReducer from "./apps";

describe("appsReducer", () => {
  let initialState: IAppState;

  const actionTypes = {
    listApps: getType(actions.apps.listApps),
    receiveAppList: getType(actions.apps.receiveAppList),
    requestApps: getType(actions.apps.requestApps),
    receiveAppUpdateInfo: getType(actions.apps.receiveAppUpdateInfo),
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
        payload: true,
        type: actionTypes.listApps as any,
      });
      expect(state).toEqual({ ...initialState, isFetching: true, listingAll: true });
      state = appsReducer(state, {
        payload: false,
        type: actionTypes.listApps as any,
      });
      expect(state).toEqual({ ...initialState, isFetching: true, listingAll: false });
    });

    describe("receieveAppUpdateInfo", () => {
      const testListOverview = [
        { releaseName: "test1" },
        { releaseName: "test2" },
        { releaseName: "test3" },
      ];
      const testUpdateInfo = {
        upToDate: false,
        chartLatestVersion: "1.0.0",
        appLatestVersion: "1.0.0",
        repository: { name: "myrepo", url: "myrepo.com" },
      };

      it("updates the listOverview entry with the updateInfo if it exists", () => {
        let state = {
          ...initialState,
          listOverview: testListOverview,
        } as IAppState;

        state = appsReducer(state, {
          payload: {
            releaseName: "test2",
            updateInfo: testUpdateInfo,
          },
          type: actionTypes.receiveAppUpdateInfo,
        });

        expect(state.listOverview![1].updateInfo).toEqual(testUpdateInfo);
      });

      it("doesn't change listOverview if the entry doesn't exist", () => {
        let state = {
          ...initialState,
          listOverview: testListOverview,
        } as IAppState;

        state = appsReducer(state, {
          payload: {
            releaseName: "test4",
            updateInfo: testUpdateInfo,
          },
          type: actionTypes.receiveAppUpdateInfo,
        });

        expect(state.listOverview).toBe(testListOverview);
      });

      it("doesn't change listOverview if it is undefined", () => {
        const state = appsReducer(initialState, {
          payload: {
            releaseName: "test4",
            updateInfo: testUpdateInfo,
          },
          type: actionTypes.receiveAppUpdateInfo,
        });

        expect(state.listOverview).toBeUndefined();
      });

      it("sets updateInfo for selected if release name matches", () => {
        const selected = { name: "test" };
        let state = {
          ...initialState,
          selected,
        } as IAppState;

        state = appsReducer(state, {
          payload: {
            releaseName: "test",
            updateInfo: testUpdateInfo,
          },
          type: actionTypes.receiveAppUpdateInfo,
        });

        expect(state.selected).toEqual({ ...selected, updateInfo: testUpdateInfo });
      });

      it("doesn't change selected if release name does not match", () => {
        const selected = { name: "test" };
        let state = {
          ...initialState,
          selected,
        } as IAppState;

        state = appsReducer(state, {
          payload: {
            releaseName: "nottest",
            updateInfo: testUpdateInfo,
          },
          type: actionTypes.receiveAppUpdateInfo,
        });

        expect(state.selected).toBe(selected);
      });

      it("doesn't set updateInfo for selected if selected is undefined", () => {
        const state = appsReducer(initialState, {
          payload: {
            releaseName: "test",
            updateInfo: testUpdateInfo,
          },
          type: actionTypes.receiveAppUpdateInfo,
        });

        expect(state.selected).toBeUndefined();
      });

      it("sets isFetching to false", () => {
        let state = {
          ...initialState,
          isFetching: true,
        } as IAppState;

        state = appsReducer(state, {
          payload: {
            releaseName: "test",
            updateInfo: testUpdateInfo,
          },
          type: actionTypes.receiveAppUpdateInfo,
        });

        expect(state.isFetching).toBe(false);
      });
    });
  });
});
