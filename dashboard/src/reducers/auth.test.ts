// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { getType } from "typesafe-actions";
import actions from "../actions";
import authReducer, { IAuthState } from "./auth";

describe("authReducer", () => {
  let initialState: IAuthState;
  const errMessage = "It's a trap";

  const actionTypes = {
    authenticating: getType(actions.auth.authenticating),
    authenticationError: getType(actions.auth.authenticationError),
    setAuthenticated: getType(actions.auth.setAuthenticated),
    setSessionExpired: getType(actions.auth.setSessionExpired),
  };

  beforeEach(() => {
    initialState = {
      sessionExpired: false,
      authenticated: false,
      authenticating: false,
      oidcAuthenticated: false,
    };
  });

  describe("initial state", () => {
    it("sets authenticated false if no key is found in local storage", () => {
      expect(authReducer(undefined, {} as any)).toEqual(initialState);
    });
  });

  // TODO(miguel) doing type assertion `as any` because in typescript 2.7 it seems
  // that the type gets lost during the creation of the map above.
  // Remove `as any` once we upgrade typescript
  describe("reducer actions", () => {
    it(`sets value passed in ${actionTypes.setAuthenticated}`, () => {
      [true, false].forEach(e => {
        expect(
          authReducer(undefined, {
            payload: { authenticated: e, oidc: false },
            type: actionTypes.setAuthenticated as any,
          }),
        ).toEqual({ ...initialState, authenticated: e });
      });
    });

    it(`resets authenticated and authenticating if type ${actionTypes.authenticating}`, () => {
      expect(
        authReducer(
          { ...initialState, authenticated: true },
          { type: actionTypes.authenticating as any },
        ),
      ).toEqual({ ...initialState, authenticating: true, authenticated: false });
    });

    it(`sets error if type ${actionTypes.authenticationError}`, () => {
      expect(
        authReducer(initialState, {
          type: actionTypes.authenticationError as any,
          payload: errMessage,
        }),
      ).toEqual({ ...initialState, authenticationError: errMessage });
    });

    it(`sets authenticating to false if type ${actionTypes.authenticationError}`, () => {
      expect(
        authReducer(
          { ...initialState, authenticating: true },
          { type: actionTypes.authenticationError as any, payload: errMessage },
        ),
      ).toEqual({ ...initialState, authenticationError: errMessage });
    });

    it("sets authenticated and oidcAuthenticated", () => {
      expect(
        authReducer(initialState, {
          type: actionTypes.setAuthenticated as any,
          payload: { authenticated: true, oidc: true },
        }),
      ).toEqual({
        ...initialState,
        authenticated: true,
        oidcAuthenticated: true,
      });
    });

    it("unsets session expired", () => {
      expect(
        authReducer(
          {
            ...initialState,
            sessionExpired: true,
          },
          {
            type: actionTypes.setSessionExpired as any,
            payload: { sessionExpired: false },
          },
        ),
      ).toEqual({
        ...initialState,
        sessionExpired: false,
      });
    });

    it("sets session expired", () => {
      expect(
        authReducer(
          {
            ...initialState,
            sessionExpired: false,
          },
          {
            payload: { sessionExpired: true },
            type: actionTypes.setSessionExpired as any,
          },
        ),
      ).toEqual({
        ...initialState,
        sessionExpired: true,
      });
    });
  });
});
