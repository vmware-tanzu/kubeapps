// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { FetchError, IInstalledPackageState } from "shared/types";
import { getType } from "typesafe-actions";
import actions from "../actions";
import installedPackagesReducer from "./installedpackages";

describe("appsReducer", () => {
  let initialState: IInstalledPackageState;

  const actionTypes = {
    requestInstalledPackageList: getType(actions.installedpackages.requestInstalledPackageList),
    requestInstalledPackage: getType(actions.installedpackages.requestInstalledPackage),
  };

  beforeEach(() => {
    initialState = {
      isFetching: false,
      items: [],
    };
  });

  describe("reducer actions", () => {
    it("sets isFetching when requesting an app", () => {
      [true, false].forEach(_e => {
        expect(
          installedPackagesReducer(undefined, {
            type: actionTypes.requestInstalledPackage as any,
          }),
        ).toEqual({ ...initialState, isFetching: true });
      });
    });

    it("toggles the listAll state", () => {
      let state = installedPackagesReducer(undefined, {
        type: actionTypes.requestInstalledPackageList as any,
      });
      expect(state).toEqual({ ...initialState, isFetching: true });
      state = installedPackagesReducer(state, {
        type: actionTypes.requestInstalledPackageList as any,
      });
      expect(state).toEqual({ ...initialState, isFetching: true });
    });

    it("clears the error after clearErrorInstalledPackage", () => {
      let state = installedPackagesReducer(undefined, {
        type: getType(actions.installedpackages.errorInstalledPackage) as any,
        payload: new FetchError("boom"),
      });
      expect(state).toEqual({ ...initialState, isFetching: false, error: new FetchError("boom") });
      state = installedPackagesReducer(state, {
        type: getType(actions.installedpackages.clearErrorInstalledPackage) as any,
      });
      expect(state).toEqual({ ...initialState, isFetching: false, error: undefined });
    });
  });
});
