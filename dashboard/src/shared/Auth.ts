// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { grpc } from "@improbable-eng/grpc-web";
import { AxiosResponse } from "axios";
import jwt from "jsonwebtoken";
import { get } from "lodash";
import { IConfig } from "./Config";
import { KubeappsGrpcClient } from "./KubeappsGrpcClient";
import {
  InternalServerNetworkError,
  NotFoundNetworkError,
  UnauthorizedNetworkError,
} from "./types";
const AuthTokenKey = "kubeapps_auth_token";
const AuthTokenOIDCKey = "kubeapps_auth_token_oidc";

export class Auth {
  public static resourcesServiceClient = (token?: string) =>
    new KubeappsGrpcClient().getResourcesServiceClientImpl(token);

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
    window.location.assign(config.oauthLogoutURI || "/oauth2/sign_out");
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
      // Trimming the b64 padding character ("=") as it is not accepted by k8s
      // https://github.com/kubernetes/apiserver/blob/release-1.22/pkg/authentication/request/websocket/protocol.go#L38
      "base64url.bearer.authorization.k8s.io." +
        Buffer.from(token).toString("base64").replaceAll("=", ""),
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
  public static async validateToken(cluster: string, token: string) {
    try {
      await this.resourcesServiceClient(token).CheckNamespaceExists({
        context: { cluster, namespace: "default" },
      });
    } catch (e: any) {
      if (e.code === grpc.Code.Unauthenticated) {
        throw new UnauthorizedNetworkError("invalid token");
      }
      // https://kubernetes.io/docs/reference/access-authn-authz/authentication/#anonymous-requests
      // Since we are always passing a token here, A 403 authorization error
      // only occurs if the token resulted in successful authentication. We
      // don't make any assumptions over RBAC for the requested namespace or
      // other required authz permissions until operations on those resources
      // are attempted (though we may want to revisit this in the future).
      if (e.code !== grpc.Code.PermissionDenied) {
        if (e.code === grpc.Code.NotFound) {
          throw new NotFoundNetworkError("not found");
        }
        if (e.code === grpc.Code.Internal) {
          throw new InternalServerNetworkError("internal error");
        }
        throw new InternalServerNetworkError(`${e.code}: ${e.message}`);
      }
    }
  }

  // isErrorFromAPIsServer returns true if the response is a 403 determined to have originated
  // from the grpc-web APIs server, rather than the auth proxy.
  public static isErrorFromAPIsServer(e: any): boolean {
    const contentType = e.metadata?.headersMap["content-type"] as string[];
    if (contentType.some(v => v.startsWith("application/grpc-web"))) {
      return true;
    }
    return false;
  }

  // is403FromAuthProxy returns true if the response is a 403 determined to have originated
  // from the auth proxy itself, rather than upstream.
  //
  // Ideally we would be able to set a header for responses generated by the
  // auth proxy, rather than rely on the fact that the 403 response sent by
  // the auth proxy is (by default) an html page (rather than the json
  // upstream result). Hence encapsulating this ugliness here so we can fix
  // it in the one spot. We may need to query `/oauth2/info` to avoid potential
  // false positives.
  // Note: This function is only used now by AxiosInstance which
  // in turn is only used for operators support.
  public static is403FromAuthProxy(r: AxiosResponse<any>): boolean {
    if (r.data && typeof r.data === "string" && r.data.match("system:serviceaccount")) {
      // If the error message is related to a service account is not from the auth proxy
      return false;
    }
    return r.status === 403 && (!r.data || !r.data.message);
  }

  // isAnonymous returns true if the message includes "system:anonymous"
  // in response.data or response.data.message
  // the k8s api server nowadays defaults to allowing anonymous
  // requests, so that rather than returning a 401, a 403 is returned if
  // RBAC does not allow the anonymous user access.
  //
  // Note: This function is only used now by AxiosInstance which
  // in turn is only used for operators support.
  public static isAnonymous(response: AxiosResponse<any>): boolean {
    const msg = get(response, "data.message") || get(response, "data");
    return typeof msg === "string" && msg.includes("system:anonymous");
  }

  // isAuthenticatedWithCookie() does a GET request to determine if
  // the request is authenticated with an http-only cookie (there is, by design,
  // no way to determine via client JS whether an http-only cookie is present).
  //
  // Note that when using the auth-proxy, anonymous requests never
  // get to the backend since the auth-proxy requires authentication.
  // But if this function is incorrectly called when the auth-proxy
  // is not in use, with a cluster that supports anonymous requests,
  // it could potentially return a false positive.
  public static async isAuthenticatedWithCookie(cluster: string): Promise<boolean> {
    try {
      await this.resourcesServiceClient().CheckNamespaceExists({
        context: { cluster, namespace: "default" },
      });
    } catch (e: any) {
      // The only error response which can possibly mean we did authenticate is
      // a 403 from the k8s api server (ie. we got through to k8s api server
      // but RBAC doesn't authorize us).
      if (e.code !== grpc.Code.PermissionDenied) {
        return false;
      }

      // A 403 error response from our APIs server, rather than the
      // auth proxy, means we are authenticated and did get
      // through to the API server but were rejected by RBAC.
      if (this.isErrorFromAPIsServer(e)) {
        return true;
      }
      return false;
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
    if (payload && payload[namespaceKey]) {
      return payload[namespaceKey];
    }
    return "";
  }
}
