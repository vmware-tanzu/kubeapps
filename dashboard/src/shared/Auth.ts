import Axios, { AxiosResponse } from "axios";
const AuthTokenKey = "kubeapps_auth_token";

export class Auth {
  public static getAuthToken() {
    return localStorage.getItem(AuthTokenKey);
  }

  public static setAuthToken(token: string) {
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

  // Throws an error if accessing the Kubernetes API returns an Unauthorized error
  public static async validate(token?: string) {
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
}

export const axios = Axios.create();
