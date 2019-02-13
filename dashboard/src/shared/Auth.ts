import Axios, { AxiosResponse } from "axios";
const AuthTokenKey = "kubeapps_auth_token";
const AuthTokenOIDCKey = "kubeapps_auth_token_oidc";

export class Auth {
  public static getAuthToken() {
    return localStorage.getItem(AuthTokenKey);
  }

  public static setAuthToken(token: string, oidc?: boolean) {
    localStorage.setItem(AuthTokenOIDCKey, (!!oidc).toString());
    return localStorage.setItem(AuthTokenKey, token);
  }

  public static unsetAuthToken() {
    return localStorage.removeItem(AuthTokenKey);
  }

  public static wsProtocols() {
    const token = this.getAuthToken();
    if (!token) {
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
      let options = {};
      if (token) {
        options = { headers: { Authorization: `Bearer ${token}` } };
      }
      await Axios.get("api/kube/", options);
    } catch (e) {
      const res = e.response as AxiosResponse;
      if (res.status === 401 || res.status === 403) {
        throw new Error("invalid token");
      }
    }
  }

  public static async fetchOIDCToken(): Promise<string | null> {
    try {
      const { headers } = await Axios.head("/");
      if (headers && headers.authorization) {
        const tokenMatch = (headers.authorization as string).match(/Bearer\s(.*)/);
        if (tokenMatch) {
          return tokenMatch[1];
        }
      }
    } catch (e) {
      // Unable to fetch
    }
    return null;
  }
}

export const axios = Axios.create();
