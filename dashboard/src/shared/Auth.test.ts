import Axios, { AxiosResponse } from "axios";
import * as jwt from "jsonwebtoken";
import { Auth } from "./Auth";
import { SupportedThemes } from "./Config";

describe("Auth", () => {
  beforeEach(() => {
    Axios.get = jest.fn();
  });
  afterEach(() => {
    jest.resetAllMocks();
  });
  it("should get an URL with the given token", async () => {
    const mock = jest.fn();
    Axios.get = mock;
    await Auth.validateToken("othercluster", "foo");
    expect(mock.mock.calls[0]).toEqual([
      "api/clusters/othercluster/",
      { headers: { Authorization: "Bearer foo" } },
    ]);
  });

  describe("when there is an error", () => {
    [
      {
        name: "should throw an invalid token error for 401 responses",
        response: { status: 401, data: "ignored anyway" },
        expectedError: new Error("invalid token"),
      },
      {
        name: "should throw a standard error for a 404 response",
        response: { status: 404, data: "Not found" },
        expectedError: new Error("404: Not found"),
      },
      {
        name: "should throw a standard error for a 500 response",
        response: { status: 500, data: "Server exception" },
        expectedError: new Error("500: Server exception"),
      },
      {
        name: "should succeed for a 403 response",
        response: { status: 403, data: "Not Allowed" },
        expectedError: null,
      },
    ].forEach(testCase => {
      it(testCase.name, async () => {
        const mock = jest.fn(() => {
          return Promise.reject({ response: testCase.response });
        });
        Axios.get = mock;
        // TODO(absoludity): tried using `expect(fn()).rejects.toThrow()` but it seems we need
        // to upgrade jest for `toThrow()` to work with async.
        let err = null;
        try {
          await Auth.validateToken("default", "foo");
        } catch (e) {
          err = e;
        } finally {
          expect(err).toEqual(testCase.expectedError);
        }
      });
    });
  });

  describe("isAuthenticatedWithCookie", () => {
    it("returns true if request to API root succeeds", async () => {
      Axios.get = jest.fn().mockReturnValue(Promise.resolve({ headers: { status: 200 } }));
      const isAuthed = await Auth.isAuthenticatedWithCookie("somecluster");

      expect(Axios.get).toBeCalledWith("api/clusters/somecluster/");
      expect(isAuthed).toBe(true);
    });
    it("returns false if the request to api root results in a 403 for an anonymous request", async () => {
      Axios.get = jest.fn(() => {
        return Promise.reject({
          response: {
            status: 403,
            data: { message: "Something with system:anonymous in there" },
          },
        });
      });
      const isAuthed = await Auth.isAuthenticatedWithCookie("somecluster");
      expect(isAuthed).toBe(false);
    });
    it("returns false if the request to api root results in a non-json response (ie. without data.message)", async () => {
      Axios.get = jest.fn(() => {
        return Promise.reject({
          response: {
            status: 403,
          },
        });
      });
      const isAuthed = await Auth.isAuthenticatedWithCookie("somecluster");
      expect(isAuthed).toBe(false);
    });
    it("returns true if the request to api root results in a 403 (but not anonymous)", async () => {
      Axios.get = jest.fn(() => {
        return Promise.reject({
          response: {
            status: 403,
            data: { message: "some message for other-user" },
          },
        });
      });
      const isAuthed = await Auth.isAuthenticatedWithCookie("somecluster");
      expect(isAuthed).toBe(true);
    });
    it("should return false if the request results in a 401", async () => {
      Axios.get = jest.fn(() => {
        return Promise.reject({ response: { status: 401 } });
      });
      const isAuthed = await Auth.isAuthenticatedWithCookie("somecluster");
      expect(isAuthed).toBe(false);
    });
  });

  describe("defaultNamespaceFromToken", () => {
    const customNamespace = "kubeapps-user";

    it("should return the k8s namespace from the jwt token", () => {
      const token = jwt.sign(
        {
          iss: "kubernetes/serviceaccount",
          "kubernetes.io/serviceaccount/namespace": customNamespace,
        },
        "secret",
      );

      const defaultNamespace = Auth.defaultNamespaceFromToken(token);

      expect(defaultNamespace).toEqual(customNamespace);
    });

    it("should return an empty string if the namespace is not present", () => {
      const token = jwt.sign(
        {
          iss: "kubernetes/serviceaccount",
        },
        "secret",
      );

      const defaultNamespace = Auth.defaultNamespaceFromToken(token);

      expect(defaultNamespace).toEqual("");
    });

    it("should return an empty string if the token cannot be decoded", () => {
      const token = "not a jwt token";

      const defaultNamespace = Auth.defaultNamespaceFromToken(token);

      expect(defaultNamespace).toEqual("");
    });
  });

  describe("unsetAuthCookie", () => {
    let mockedAssign: jest.Mocked<(url: string) => void>;
    beforeEach(() => {
      mockedAssign = jest.fn();
      // After the JSDOM upgrade, window.xxx are read-only properties
      // https://github.com/facebook/jest/issues/9471
      Object.defineProperty(window, "location", {
        configurable: true,
        writable: true,
        value: { assign: mockedAssign },
      });
    });

    it("uses the config to redirect to a logout URL", () => {
      const oauthLogoutURI = "/example/logout";

      Auth.unsetAuthCookie({
        oauthLoginURI: "",
        authProxyEnabled: true,
        oauthLogoutURI,
        kubeappsCluster: "default",
        kubeappsNamespace: "ns",
        appVersion: "2",
        clusters: [],
        authProxySkipLoginPage: false,
        theme: SupportedThemes.light,
      });

      expect(mockedAssign).toBeCalledWith(oauthLogoutURI);
      expect(localStorage.removeItem).toBeCalled();
    });

    it("defaults to the oauth2-proxy logout URI", () => {
      Auth.unsetAuthCookie({
        oauthLoginURI: "",
        authProxyEnabled: true,
        oauthLogoutURI: "",
        kubeappsCluster: "default",
        kubeappsNamespace: "ns",
        appVersion: "2",
        clusters: [],
        authProxySkipLoginPage: false,
        theme: SupportedThemes.light,
      });

      expect(mockedAssign).toBeCalledWith("/oauth2/sign_out");
    });
  });
});

describe("is403FromAuthProxy", () => {
  it("does not assume to be authenticated from the auth proxy if the message contains a service account", () => {
    expect(
      Auth.is403FromAuthProxy({
        status: 403,
        data:
          'namespaces is forbidden: User "system:serviceaccount:kubeapps:kubeapps-internal-kubeops" cannot list resource "namespaces" in API group "" at the cluster scope',
      } as AxiosResponse<any>),
    ).toBe(false);
  });
});

describe("isAnonymous", () => {
  it("returns true if the message includes 'system:anonymous' in response.data", () => {
    expect(
      Auth.isAnonymous({
        status: 403,
        data:
          '{"metadata":{},"status":"Failure","message":"selfsubjectaccessreviews.authorization.k8s.io is forbidden: User "system:anonymous" cannot create resource "selfsubjectaccessreviews" in API group "authorization.k8s.io" at the cluster scope","reason":"Forbidden","details":{"group":"authorization.k8s.io","kind":"selfsubjectaccessreviews"},"code":403} {"namespaces":null}',
      } as AxiosResponse<any>),
    ).toBe(true);
  });
  it("returns true if the message includes 'system:anonymous' in response.data.message", () => {
    expect(
      Auth.isAnonymous({
        status: 403,
        data: {
          message:
            '{"metadata":{},"status":"Failure","message":"selfsubjectaccessreviews.authorization.k8s.io is forbidden: User "system:anonymous" cannot create resource "selfsubjectaccessreviews" in API group "authorization.k8s.io" at the cluster scope","reason":"Forbidden","details":{"group":"authorization.k8s.io","kind":"selfsubjectaccessreviews"},"code":403} {"namespaces":null}',
        },
      } as AxiosResponse<any>),
    ).toBe(true);
  });
  it("returns false if the message does not include 'system:anonymous' in response.data", () => {
    expect(
      Auth.isAnonymous({
        status: 403,
        data:
          'namespaces is forbidden: User "system:serviceaccount:kubeapps:kubeapps-internal-kubeops" cannot list resource "namespaces" in API group "" at the cluster scope',
      } as AxiosResponse<any>),
    ).toBe(false);
  });
  it("returns false if the message does not include 'system:anonymous' in response.data.message", () => {
    expect(
      Auth.isAnonymous({
        status: 403,
        data: {
          message:
            'namespaces is forbidden: User "system:serviceaccount:kubeapps:kubeapps-internal-kubeops" cannot list resource "namespaces" in API group "" at the cluster scope',
        },
      } as AxiosResponse<any>),
    ).toBe(false);
  });
});
