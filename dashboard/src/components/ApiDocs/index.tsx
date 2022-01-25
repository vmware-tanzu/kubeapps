// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { Auth } from "shared/Auth";
import SwaggerUI from "swagger-ui-react";
import "swagger-ui-react/swagger-ui.css";
import "./ApiDocs.css";

// Request interface needed for avoiding type error in the requestInterceptor
// it is being used, but it is not exported, so we define it here
interface Request {
  [k: string]: any;
}

function authenticate(req: Request): Request {
  const token = Auth.getAuthToken() || "";
  // Otherwise, OIDC login is being used and we cannot automatically set the token here
  if (token) {
    req.headers.Authorization = `Bearer ${token}`;
  }
  req.url = req.url.replace("127.0.0.1:8080", `${window.location.host}`);
  return req;
}

export default function ApiDocs() {
  return <SwaggerUI url="/openapi.yaml" docExpansion="list" requestInterceptor={authenticate} />;
}
