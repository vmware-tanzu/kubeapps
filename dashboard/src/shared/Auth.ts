import Axios, { AxiosRequestConfig, AxiosResponse } from "axios";

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

  // Throws an error if the token is invalid
  public static async validateToken(token: string) {
    try {
      await Axios.get("/api/kube/", { headers: { Authorization: `Bearer ${token}` } });
    } catch (e) {
      const res = e.response as AxiosResponse;
      if (res.status === 401) {
        throw new Error("invalid token");
      }
    }
  }
}

// authenticatedAxiosInstance returns an axios instance with an interceptor
// configured to set the current auth token
function authenticatedAxiosInstance() {
  const a = Axios.create();
  a.interceptors.request.use((config: AxiosRequestConfig) => {
    const authToken = Auth.getAuthToken();
    if (authToken) {
      config.headers.Authorization = `Bearer ${authToken}`;
    }
    return config;
  });
  return a;
}

export const axios = authenticatedAxiosInstance();
