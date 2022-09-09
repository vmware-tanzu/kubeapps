// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import * as moxios from "moxios";
import { IAuthState } from "reducers/auth";
import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";
import { addAuthHeaders, addErrorHandling, axios } from "shared/AxiosInstance";
import { Auth } from "./Auth";
import { initialState } from "./specs/mountWrapper";
import {
  ConflictNetworkError,
  ForbiddenNetworkError,
  InternalServerNetworkError,
  IStoreState,
  NotFoundNetworkError,
  UnauthorizedNetworkError,
  UnprocessableEntityError,
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
        ...initialState.auth,
        state,
      },
    } as Partial<IStoreState>);

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
    { code: 401, errorClass: UnauthorizedNetworkError },
    { code: 403, errorClass: ForbiddenNetworkError },
    { code: 404, errorClass: NotFoundNetworkError },
    { code: 409, errorClass: ConflictNetworkError },
    { code: 422, errorClass: UnprocessableEntityError },
    { code: 500, errorClass: InternalServerNetworkError },
  ];

  testCases.forEach(t => {
    it(`returns a custom message if ${t.code} returned`, async () => {
      moxios.stubRequest(testPath, {
        response: { message: `Will raise ${t.errorClass.name}` },
        status: t.code,
      });
      await expect(axios.get(testPath)).rejects.toThrow(`Will raise ${t.errorClass.name}`);
    });

    it(`returns the custom error ${t.errorClass.name} if ${t.code} returned`, async () => {
      moxios.stubRequest(testPath, {
        response: {},
        status: t.code,
      });
      await expect(axios.get(testPath)).rejects.toThrowError(t.errorClass);
    });
  });

  it("returns the generic error message otherwise", async () => {
    moxios.stubRequest(testPath, {
      response: {},
      status: 555,
    });
    await expect(axios.get(testPath)).rejects.toThrow("Request failed with status code 555");
  });

  it("returns the response message", async () => {
    moxios.stubRequest(testPath, {
      response: { message: "this is an error!" },
      status: 555,
    });
    await expect(axios.get(testPath)).rejects.toThrow("this is an error!");
  });

  it("dispatches auth error and logout if 401 with auth proxy", async () => {
    Auth.usingOIDCToken = jest.fn().mockReturnValue(true);
    Auth.unsetAuthCookie = jest.fn();
    const expectedActions = [
      {
        payload: "Boom!",
        type: "AUTHENTICATION_ERROR",
      },
      {
        payload: { sessionExpired: true },
        type: "SET_AUTHENTICATION_SESSION_EXPIRED",
      },
    ];

    moxios.stubRequest(testPath, {
      response: { message: "Boom!" },
      status: 401,
    });
    await expect(axios.get(testPath)).rejects.toThrow("Boom!");
    expect(store.getActions()).toEqual(expectedActions);
    expect(Auth.unsetAuthCookie).toHaveBeenCalled();
  });

  it("dispatches auth error and logout if 403 with auth proxy", async () => {
    Auth.usingOIDCToken = jest.fn().mockReturnValue(true);
    Auth.unsetAuthCookie = jest.fn();
    const expectedActions = [
      {
        payload: "not ajson paylod",
        type: "AUTHENTICATION_ERROR",
      },
      {
        payload: { sessionExpired: true },
        type: "SET_AUTHENTICATION_SESSION_EXPIRED",
      },
    ];

    moxios.stubRequest(testPath, {
      responseText: "not ajson paylod",
      status: 401,
    });
    await expect(axios.get(testPath)).rejects.toThrow("not ajson paylod");
    expect(store.getActions()).toEqual(expectedActions);
    expect(Auth.unsetAuthCookie).toHaveBeenCalled();
  });

  it("dispatches auth error and logout if 403 anonymous user and no auth proxy", async () => {
    Auth.usingOIDCToken = jest.fn().mockReturnValue(false);
    Auth.unsetAuthToken = jest.fn();
    const expectedActions = [
      {
        type: "AUTHENTICATION_ERROR",
        payload:
          '{"metadata":{},"status":"Failure","message":"selfsubjectaccessreviews.authorization.k8s.io is forbidden: User "system:anonymous" cannot create resource "selfsubjectaccessreviews" in API group "authorization.k8s.io" at the cluster scope","reason":"Forbidden","details":{"group":"authorization.k8s.io","kind":"selfsubjectaccessreviews"},"code":403} {"namespaces":null}',
      },
      {
        type: "SET_AUTHENTICATED",
        payload: {
          authenticated: false,
          oidc: false,
        },
      },
      {
        type: "CLEAR_CLUSTERS",
      },
    ];

    moxios.stubRequest(testPath, {
      response: {
        message:
          '{"metadata":{},"status":"Failure","message":"selfsubjectaccessreviews.authorization.k8s.io is forbidden: User "system:anonymous" cannot create resource "selfsubjectaccessreviews" in API group "authorization.k8s.io" at the cluster scope","reason":"Forbidden","details":{"group":"authorization.k8s.io","kind":"selfsubjectaccessreviews"},"code":403} {"namespaces":null}',
      },
      status: 403,
    });
    await expect(axios.get(testPath)).rejects.toThrow(
      '{"metadata":{},"status":"Failure","message":"selfsubjectaccessreviews.authorization.k8s.io is forbidden: User "system:anonymous" cannot create resource "selfsubjectaccessreviews" in API group "authorization.k8s.io" at the cluster scope","reason":"Forbidden","details":{"group":"authorization.k8s.io","kind":"selfsubjectaccessreviews"},"code":403} {"namespaces":null}',
    );
    expect(store.getActions()).toEqual(expectedActions);
    expect(Auth.unsetAuthToken).toHaveBeenCalled();
  });

  it("parses a forbidden response", async () => {
    moxios.stubRequest(testPath, {
      response: {
        message:
          '[{"apiGroup": "v1", "resource": "secrets", "namespace": "default", "verbs": ["list", "get"]}]',
      },
      status: 403,
    });
    await expect(axios.get(testPath)).rejects.toThrow(
      'Forbidden error, missing permissions: apiGroup: "v1", resource: "secrets", action: "list, get", namespace: default',
    );
  });
});
