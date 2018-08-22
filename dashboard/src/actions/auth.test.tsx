import { IAuthState } from "reducers/auth";
import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";
import { IStoreState } from "shared/types";
import { getType } from "typesafe-actions";
import actions from ".";
import { IStoreState } from "../../shared/types";
import { Auth } from "../shared/Auth";

const mockStore = configureMockStore([thunk]);
const token = "abcd";
const validationErrorMsg = "Validation error";

let store: IStoreState;

beforeEach(() => {
  const state: IAuthState = {
    authenticated: false,
    authenticating: false,
  };

  Auth.validateToken = jest.fn();
  Auth.setAuthToken = jest.fn();

  store = mockStore({
    auth: {
      state,
    },
  });
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
        errorMsg: `Error: ${validationErrorMsg}`,
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
        authenticated: true,
        type: getType(actions.auth.setAuthenticated),
      },
    ];

    return store.dispatch(actions.auth.authenticate(token)).then(() => {
      expect(store.getActions()).toEqual(expectedActions);
    });
  });
});
