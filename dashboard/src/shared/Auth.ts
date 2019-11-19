import Axios, { AxiosResponse } from "axios";
import * as jwt from "jsonwebtoken";
const AuthTokenKey = "kubeapps_auth_token";
const AuthTokenOIDCKey = "kubeapps_auth_token_oidc";
import { IConfig } from "./Config";
import { APIBase } from "./Kube";

export const DEFAULT_NAMESPACE = "default";

export class Auth {
  public static getAuthToken() {
    return localStorage.getItem(AuthTokenKey);
  }

  public static setAuthToken(token: string, oidc: boolean) {
    localStorage.setItem(AuthTokenOIDCKey, oidc.toString());
    if (token) {
      localStorage.setItem(AuthTokenKey, token);
    }
  }

  public static unsetAuthToken() {
    localStorage.removeItem(AuthTokenKey);
  }

  public static unsetAuthCookie(config: IConfig) {
    // http cookies cannot be deleted (or modified or read) from client-side
    // JS, so force browser to load the sign-out URI (which expires the
    // session cookie).
    localStorage.removeItem(AuthTokenOIDCKey);
    document.location.assign(config.logoutURI);
  }

  public static usingOIDCToken() {
    return localStorage.getItem(AuthTokenOIDCKey) === "true";
  }

  public static wsProtocols() {
    const token = this.getAuthToken();
    // If we're using OIDC for auth, then let the auth proxy handle
    // injecting the ws creds.
    if (!token || this.usingOIDCToken()) {
      return [];
    }
    return [
      "base64url.bearer.authorization.k8s.io." + btoa(token).replace(/=*$/g, ""),
      "binary.k8s.io",
    ];
  }

  public static fetchOptions(): RequestInit {
    const headers = new Headers();
    headers.append("Authorization", `Bearer ${this.getAuthToken()}`);
    return {
      headers,
    };
  }

  // Throws an error if the token is invalid
  public static async validateToken(token: string) {
    try {
      await Axios.get(APIBase + "/", { headers: { Authorization: `Bearer ${token}` } });
    } catch (e) {
      const res = e.response as AxiosResponse;
      if (res.status === 401) {
        throw new Error("invalid token");
      }
      // A 403 authorization error only occurs if the token resulted in
      // successful authentication. We don't make any assumptions over RBAC
      // for the root "/" nonResourceURL or other required authz permissions
      // until operations on those resources are attempted (though we may
      // want to revisit this in the future).
      if (res.status !== 403) {
        throw new Error(`${res.status}: ${res.data}`);
      }
    }
  }

  // isAuthenticatedWithCookie() does a HEAD request (without any token obviously)
  // to determine if the request is authenticated (ie. not a 401). Unfortunately
  // Kubernetes defaulting to allow anonymous requests means that this will be a 403
  // even if there are no credentials, so we additionally check the message and assume
  // we are authenticated with a cookie if a 403 is not for an anon user.
  public static async isAuthenticatedWithCookie(): Promise<boolean> {
    try {
      await Axios.get(APIBase + "/");
    } catch (e) {
      const response = e.response as AxiosResponse;
      const isAnon =
        response.data &&
        response.data.message &&
        response.data.message.includes("system:anonymous");
      return response.status === 403 && !isAnon;
    }
    return true;
  }

  // defaultNamespaceFromToken decodes a jwt token to return the k8s service
  // account namespace.
  // TODO(mnelson): until we call jwt.verify on the token during validateToken above
  // we use a default namespace for both invalid tokens and tokens without the expected
  // key.
  public static defaultNamespaceFromToken(token: string) {
    const payload = jwt.decode(token);
    const namespaceKey = "kubernetes.io/serviceaccount/namespace";
    if (!payload || !payload[namespaceKey]) {
      return DEFAULT_NAMESPACE;
    }
    return payload[namespaceKey];
  }
}
