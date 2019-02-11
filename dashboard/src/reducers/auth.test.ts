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
    setAutoAuthenticated: getType(actions.auth.setAutoAuthenticated),
  };

  beforeEach(() => {
    initialState = {
      authenticated: false,
      authenticating: false,
      autoAuthenticating: false,
      autoAuthenticated: false,
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
            payload: e,
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

    it(`resets authenticated, authenticating and sets error if type ${
      actionTypes.authenticationError
    }`, () => {
      expect(
        authReducer(
          {
            authenticating: true,
            authenticated: true,
            autoAuthenticating: true,
            autoAuthenticated: true,
          },
          { type: actionTypes.authenticationError as any, payload: errMessage },
        ),
      ).toEqual({ ...initialState, authenticationError: errMessage });
    });

    it("sets authenticated and autoauthenticated", () => {
      expect(
        authReducer(
          {
            authenticating: true,
            authenticated: false,
            autoAuthenticating: true,
            autoAuthenticated: false,
          },
          { type: actionTypes.setAutoAuthenticated as any, payload: true },
        ),
      ).toEqual({
        authenticating: false,
        authenticated: true,
        autoAuthenticating: false,
        autoAuthenticated: true,
      });
    });
  });
});
