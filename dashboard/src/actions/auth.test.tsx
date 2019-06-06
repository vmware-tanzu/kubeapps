import { IAuthState } from "reducers/auth";
import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";
import { getType } from "typesafe-actions";
import actions from ".";
import { Auth } from "../shared/Auth";

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
  Auth.fetchOIDCToken = jest.fn(() => "token");
  Auth.setAuthToken = jest.fn();
  Auth.unsetAuthToken = jest.fn();

  store = mockStore({
    auth: {
      state,
    },
  });
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

    return store.dispatch(actions.auth.authenticate(token, false)).then(() => {
      expect(store.getActions()).toEqual(expectedActions);
    });
  });

  it("dispatches authenticating and auth ok if valid", () => {
    const expectedActions = [
      {
        type: getType(actions.auth.authenticating),
      },
      {
        payload: { authenticated: true, oidc: false },
        type: getType(actions.auth.setAuthenticated),
      },
    ];

    return store.dispatch(actions.auth.authenticate(token, false)).then(() => {
      expect(store.getActions()).toEqual(expectedActions);
    });
  });
});

describe("OIDC authentication", () => {
  it("dispatches authenticating and auth error if invalid", () => {
    Auth.validateToken = jest.fn().mockImplementationOnce(() => {
      throw new Error(validationErrorMsg);
    });
    const expectedActions = [
      {
        type: getType(actions.auth.authenticating),
      },
      {
        type: getType(actions.auth.authenticating),
      },
      {
        payload: `Error: ${validationErrorMsg}`,
        type: getType(actions.auth.authenticationError),
      },
    ];

    return store.dispatch(actions.auth.tryToAuthenticateWithOIDC()).then(() => {
      expect(store.getActions()).toEqual(expectedActions);
    });
  });

  it("dispatches authenticating and auth ok if valid", () => {
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

    return store.dispatch(actions.auth.tryToAuthenticateWithOIDC()).then(() => {
      expect(store.getActions()).toEqual(expectedActions);
    });
  });

  it("expires the session and logs out ", () => {
    Auth.usingOIDCToken = jest.fn(() => true);
    const expectedActions = [
      {
        payload: { sessionExpired: true },
        type: getType(actions.auth.setSessionExpired),
      },
      {
        payload: { authenticated: false, oidc: false },
        type: getType(actions.auth.setAuthenticated),
      },
    ];

    return store.dispatch(actions.auth.expireSession()).then(() => {
      expect(store.getActions()).toEqual(expectedActions);
    });
  });
});
