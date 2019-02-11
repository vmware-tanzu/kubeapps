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
    authenticated: false,
    authenticating: false,
    autoAuthenticating: false,
    autoAuthenticated: false,
  };

  Auth.validateToken = jest.fn();
  Auth.setAuthToken = jest.fn();

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

    return store.dispatch(actions.auth.authenticate(token)).then(() => {
      expect(store.getActions()).toEqual(expectedActions);
    });
  });

  it("dispatches authenticating and auth ok if valid", () => {
    const expectedActions = [
      {
        type: getType(actions.auth.authenticating),
      },
      {
        payload: true,
        type: getType(actions.auth.setAuthenticated),
      },
    ];

    return store.dispatch(actions.auth.authenticate(token)).then(() => {
      expect(store.getActions()).toEqual(expectedActions);
    });
  });
});

describe("auto authenticate", () => {
  it("dispatches authenticating and auth error if invalid", () => {
    Auth.fetchToken = jest.fn().mockImplementationOnce(() => {
      throw new Error(validationErrorMsg);
    });
    const expectedActions = [
      {
        type: getType(actions.auth.autoAuthenticating),
      },
      {
        payload: `Error: ${validationErrorMsg}`,
        type: getType(actions.auth.authenticationError),
      },
    ];

    return store.dispatch(actions.auth.tryToAutoAuthenticate()).then(() => {
      expect(store.getActions()).toEqual(expectedActions);
    });
  });

  it("dispatches authenticating but no auth if doesn't return a token", () => {
    Auth.fetchToken = jest.fn().mockImplementationOnce(() => {
      return null;
    });

    const expectedActions = [
      {
        type: getType(actions.auth.autoAuthenticating),
      },
      {
        payload: false,
        type: getType(actions.auth.setAutoAuthenticated),
      },
    ];

    return store.dispatch(actions.auth.tryToAutoAuthenticate()).then(() => {
      expect(store.getActions()).toEqual(expectedActions);
    });
  });

  it("dispatches authenticating and auth ok if valid", () => {
    Auth.fetchToken = jest.fn().mockImplementationOnce(() => {
      return "token";
    });

    const expectedActions = [
      {
        type: getType(actions.auth.autoAuthenticating),
      },
      {
        payload: true,
        type: getType(actions.auth.setAutoAuthenticated),
      },
    ];

    return store.dispatch(actions.auth.tryToAutoAuthenticate()).then(() => {
      expect(store.getActions()).toEqual(expectedActions);
    });
  });
});
