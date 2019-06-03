import * as moxios from "moxios";
import { IAuthState } from "reducers/auth";
import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";
import { addAuthHeaders, addErrorHandling, axios } from "../shared/AxiosInstance";
import { Auth } from "./Auth";
import {
  ConflictError,
  ForbiddenError,
  NotFoundError,
  UnauthorizedError,
  UnprocessableEntity,
} from "./types";

describe("createAxiosInterceptorWithAuth", () => {
  const mockStore = configureMockStore([thunk]);
  const testPath = "/internet-is-in-a-box";
  const authToken = "search-google-in-google";

  let store: any;

  beforeAll(() => {
    const state: IAuthState = {
      sessionExpired: false,
      authenticated: false,
      authenticating: false,
      oidcAuthenticated: false,
    };

    store = mockStore({
      auth: {
        state,
      },
    });

    Auth.validateToken = jest.fn();
    Auth.setAuthToken = jest.fn();
    Auth.unsetAuthToken = jest.fn();

    Auth.getAuthToken = jest.fn().mockImplementationOnce(() => {
      return authToken;
    });

    addErrorHandling(axios, store);
    addAuthHeaders(axios);
  });

  beforeEach(() => {
    // Import as "any" to avoid typescript syntax error
    moxios.install(axios as any);
  });

  afterEach(() => {
    moxios.uninstall(axios as any);
    store.clearActions();
  });

  it("includes the auth token if provided", async () => {
    moxios.stubRequest(testPath, {});

    await axios.get(testPath);
    const request = moxios.requests.mostRecent();
    expect(request.headers.Authorization).toBe(`Bearer ${authToken}`);
  });

  const testCases = [
    { code: 401, errorClass: UnauthorizedError },
    { code: 403, errorClass: ForbiddenError },
    { code: 404, errorClass: NotFoundError },
    { code: 409, errorClass: ConflictError },
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

    it(`returns the custom error ${t.errorClass.name} if ${t.code} returned`, async () => {
      moxios.stubRequest(testPath, {
        response: {},
        status: t.code,
      });

      try {
        await axios.get(testPath);
      } catch (error) {
        expect(error.constructor).toBe(t.errorClass);
      }
    });
  });

  it("returns the generic error message otherwise", async () => {
    moxios.stubRequest(testPath, {
      response: {},
      status: 555,
    });

    try {
      await axios.get(testPath);
    } catch (error) {
      expect(error.message).toBe("Request failed with status code 555");
    }
  });

  it("returns the response message", async () => {
    moxios.stubRequest(testPath, {
      response: { message: "this is an error!" },
      status: 555,
    });

    try {
      await axios.get(testPath);
    } catch (error) {
      expect(error.message).toBe("this is an error!");
    }
  });

  it("dispatches auth error if 401", async () => {
    const expectedActions = [
      {
        payload: "Boom!",
        type: "AUTHENTICATION_ERROR",
      },
      {
        payload: { authenticated: false, oidc: false },
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
