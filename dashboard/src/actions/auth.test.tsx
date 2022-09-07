// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { IAuthState } from "reducers/auth";
import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";
import { Auth } from "shared/Auth";
import Namespace, * as NS from "shared/Namespace";
import { initialState } from "shared/specs/mountWrapper";
import { IStoreState } from "shared/types";
import { getType } from "typesafe-actions";
import actions from ".";

const defaultCluster = "default";
const mockStore = configureMockStore([thunk]);
const token = "abcd";
const validationErrorMsg = "Validation error";

let store: any;

beforeEach(() => {
  const state: IAuthState = {
    sessionExpired: false,
    authenticated: false,
    authenticating: false,
    oidcAuthenticated: false,
  };

  Auth.validateToken = jest.fn();
  Auth.isAuthenticatedWithCookie = jest.fn().mockReturnValue("token");
  Auth.setAuthToken = jest.fn();
  Auth.unsetAuthToken = jest.fn();
  Namespace.list = jest.fn(async () => []);
  jest.spyOn(NS, "unsetStoredNamespace");

  store = mockStore({
    auth: {
      ...initialState.auth,
      state,
    },
    config: {
      ...initialState.config,
      oauthLogoutURI: "/log/out",
    },
  } as Partial<IStoreState>);
});

afterEach(() => {
  jest.clearAllMocks();
});

describe("authenticate", () => {
  it("dispatches authenticating and auth error if invalid", () => {
    Auth.validateToken = jest.fn().mockImplementationOnce(() => {
      throw new Error(validationErrorMsg);
    });
    const expectedActions = [
      {
        type: getType(actions.auth.authenticating),
      },
      {
        payload: `Error: ${validationErrorMsg}`,
        type: getType(actions.auth.authenticationError),
      },
    ];

    return store.dispatch(actions.auth.authenticate("default", token, false)).then(() => {
      expect(store.getActions()).toEqual(expectedActions);
    });
  });

  it("dispatches authenticating and auth ok if valid", () => {
    Auth.validateToken = jest.fn();
    const expectedActions = [
      {
        type: getType(actions.auth.authenticating),
      },
      {
        payload: { authenticated: true, oidc: false },
        type: getType(actions.auth.setAuthenticated),
      },
    ];

    return store.dispatch(actions.auth.authenticate(defaultCluster, token, false)).then(() => {
      expect(store.getActions()).toEqual(expectedActions);
      expect(Auth.validateToken).toHaveBeenCalledWith(defaultCluster, token);
    });
  });

  it("does not validate a token if oidc is true", () => {
    Auth.validateToken = jest.fn();
    const expectedActions = [
      {
        type: getType(actions.auth.authenticating),
      },
      {
        payload: { authenticated: true, oidc: true },
        type: getType(actions.auth.setAuthenticated),
      },
      {
        payload: { sessionExpired: false },
        type: getType(actions.auth.setSessionExpired),
      },
    ];

    return store.dispatch(actions.auth.authenticate(defaultCluster, "ignored", true)).then(() => {
      expect(store.getActions()).toEqual(expectedActions);
      expect(Auth.validateToken).not.toHaveBeenCalled();
    });
  });
});

describe("OIDC authentication", () => {
  it("dispatches authenticating and auth ok if valid", async () => {
    Auth.isAuthenticatedWithCookie = jest.fn().mockReturnValue(true);
    const expectedActions = [
      {
        type: getType(actions.auth.authenticating),
      },
      {
        type: getType(actions.auth.authenticating),
      },
      {
        payload: { authenticated: true, oidc: true },
        type: getType(actions.auth.setAuthenticated),
      },
      {
        payload: { sessionExpired: false },
        type: getType(actions.auth.setSessionExpired),
      },
    ];

    return store.dispatch(actions.auth.checkCookieAuthentication("default")).then(() => {
      expect(store.getActions()).toEqual(expectedActions);
    });
  });

  it("expires the session and logs out", () => {
    Auth.usingOIDCToken = jest.fn(() => true);
    // After the JSDOM upgrade, window.xxx are read-only properties
    // https://github.com/facebook/jest/issues/9471
    Object.defineProperty(window, "location", {
      configurable: true,
      writable: true,
      value: { assign: jest.fn() },
    });
    const expectedActions = [
      {
        payload: { sessionExpired: true },
        type: getType(actions.auth.setSessionExpired),
      },
    ];

    return store.dispatch(actions.auth.expireSession()).then(() => {
      expect(store.getActions()).toEqual(expectedActions);
      expect(localStorage.removeItem).toBeCalled();
      expect(window.location.assign).toBeCalledWith("/log/out");
    });
  });
});

describe("logout", () => {
  it("unsets the stored namespace", async () => {
    await store.dispatch(actions.auth.logout());
    expect(NS.unsetStoredNamespace).toHaveBeenCalled();
  });
});
