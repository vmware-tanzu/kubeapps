import Axios, { AxiosRequestConfig } from "axios";

const AuthTokenKey = "kubeapps_auth_token";

export class Auth {
  public static getAuthToken() {
    return localStorage.getItem(AuthTokenKey);
  }

  public static setAuthToken(token: string) {
    return localStorage.setItem(AuthTokenKey, token);
  }

  public static wsProtocols() {
    const token = this.getAuthToken();
    if (!token) {
      return [];
    }
    return ["base64url.bearer.authorization.k8s.io." + btoa(token).replace(/=*$/g, "")];
  }

  public static fetchOptions() {
    return { headers: { Authorization: `Bearer ${this.getAuthToken()}` } };
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
