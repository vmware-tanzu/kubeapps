import { AxiosInstance } from "axios";
import * as moxios from "moxios";
import { IAuthState } from "reducers/auth";
import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";
import { Auth, createAxiosInstance } from "./Auth";
import {
  AppConflict,
  ForbiddenError,
  NotFoundError,
  UnauthorizedError,
  UnprocessableEntity,
} from "./types";

describe("authenticatedAxiosInstance", () => {
  const mockStore = configureMockStore([thunk]);
  const testPath = "/internet-is-in-a-box";
  const authToken = "search-google-in-google";

  let store: any;
  let axios: AxiosInstance;

  beforeEach(() => {
    const state: IAuthState = {
      authenticated: false,
      authenticating: false,
    };

    Auth.validateToken = jest.fn();
    Auth.setAuthToken = jest.fn();
    Auth.unsetAuthToken = jest.fn();

    Auth.getAuthToken = jest.fn().mockImplementationOnce(() => {
      return authToken;
    });

    store = mockStore({
      auth: {
        state,
      },
    });
    axios = createAxiosInstance(store);
    moxios.install(axios);
  });

  afterEach(() => {
    moxios.uninstall(axios);
  });

  it("it includes the auth token if provided", async () => {
    moxios.stubRequest(testPath, {});

    await axios.get(testPath);
    const request = moxios.requests.mostRecent();
    expect(request.headers.Authorization).toBe(`Bearer ${authToken}`);
  });

  const testCases = [
    { code: 401, errorClass: UnauthorizedError },
    { code: 403, errorClass: ForbiddenError },
    { code: 404, errorClass: NotFoundError },
    { code: 409, errorClass: AppConflict },
    { code: 422, errorClass: UnprocessableEntity },
  ];

  testCases.forEach(t => {
    it(`returns a custom message if ${t.code} returned`, async () => {
      moxios.stubRequest(testPath, {
        response: { message: `Will raise ${t.errorClass.name}` },
        status: t.code,
      });

      try {
        await axios.get(testPath);
      } catch (error) {
        expect(error.message).toBe(`Will raise ${t.errorClass.name}`);
      }
    });
  });

  it(`returns the generic error message otherwise`, async () => {
    moxios.stubRequest(testPath, {
      response: { message: "this will be ignored" },
      status: 555,
    });

    try {
      await axios.get(testPath);
    } catch (error) {
      expect(error.message).toBe("Request failed with status code 555");
    }
  });

  it("it dispatches auth error and logout if 401", async () => {
    const expectedActions = [
      {
        errorMsg: "Boom!",
        type: "AUTHENTICATION_ERROR",
      },
      {
        authenticated: false,
        type: "SET_AUTHENTICATED",
      },
    ];

    moxios.stubRequest(testPath, {
      response: { message: "Boom!" },
      status: 401,
    });

    try {
      await axios.get(testPath);
    } catch (error) {
      expect(store.getActions()).toEqual(expectedActions);
    }
  });
});
