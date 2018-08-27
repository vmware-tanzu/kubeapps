import { getType } from "typesafe-actions";
import actions from "../actions";
import authReducer, { IAuthState } from "./auth";

describe("authReducer", () => {
  let initialState: IAuthState;
  const errMessage = "It's a trap";

  beforeEach(() => {
    initialState = {
      authenticated: false,
      authenticating: false,
    };
  });

  describe("initial state", () => {
    it("sets authenticated false if no key is found in local storage", () => {
      expect(authReducer(undefined, { type: "NONE" })).toEqual(initialState);
    });
  });

  describe("reducer actions", () => {
    it("sets value passed in setAuthenticated", () => {
      [true, false].forEach(e => {
        expect(
          authReducer(undefined, {
            type: getType(actions.auth.setAuthenticated),
            authenticated: e,
          }),
        ).toEqual({ ...initialState, authenticated: e });
      });
    });

    it("resets authenticated and authenticating if type: AUTHENTICATING", () => {
      expect(
        authReducer(
          { ...initialState, authenticated: true },
          { type: getType(actions.auth.authenticating) },
        ),
      ).toEqual({ ...initialState, authenticating: true, authenticated: false });
    });

    it("resets authenticated, authenticating and sets error if type: AUTHENTICATION_ERROR", () => {
      expect(
        authReducer(
          { authenticating: true, authenticated: true },
          { type: getType(actions.auth.authenticationError), errorMsg: errMessage },
        ),
      ).toEqual({ ...initialState, authenticationError: errMessage });
    });
  });
});
