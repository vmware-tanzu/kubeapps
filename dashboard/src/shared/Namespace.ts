// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { get } from "lodash";
import { Auth } from "./Auth";
import { KubeappsGrpcClient } from "./KubeappsGrpcClient";
import { convertGrpcAuthError } from "./utils";

export default class Namespace {
  private static resourcesServiceClient = () =>
    new KubeappsGrpcClient().getResourcesServiceClientImpl();

  public static async list(cluster: string) {
    const { namespaceNames } = await this.resourcesServiceClient()
      .GetNamespaceNames({ cluster: cluster })
      .catch((e: any) => {
        throw convertGrpcAuthError(e);
      });
    return namespaceNames;
  }

  public static async create(
    cluster: string,
    namespace: string,
    labels: { [key: string]: string },
  ) {
    await this.resourcesServiceClient()
      .CreateNamespace({
        context: {
          cluster,
          namespace,
        },
        labels: labels,
      })
      .catch((e: any) => {
        throw convertGrpcAuthError(e);
      });
  }

  public static async exists(cluster: string, namespace: string) {
    const { exists } = await this.resourcesServiceClient()
      .CheckNamespaceExists({
        context: {
          cluster,
          namespace,
        },
      })
      .catch((e: any) => {
        throw convertGrpcAuthError(e);
      });
    return exists;
  }
}

// The namespace information will contain a map[cluster]:namespace with the default namespaces
const namespaceKey = "kubeapps_namespace";

function parseStoredNS() {
  const ns = localStorage.getItem(namespaceKey) || "{}";
  let parsedNS = {};
  try {
    parsedNS = JSON.parse(ns);
  } catch (e: any) {
    // The stored value should be a json object, if not, ignore it
  }
  return parsedNS;
}

function getStoredNamespace(cluster: string) {
  return get(parseStoredNS(), cluster, "");
}

export function setStoredNamespace(cluster: string, namespace: string) {
  const ns = parseStoredNS();
  ns[cluster] = namespace;
  localStorage.setItem(namespaceKey, JSON.stringify(ns));
}

export function unsetStoredNamespace() {
  localStorage.removeItem(namespaceKey);
}

export function getCurrentNamespace(cluster: string, currentNS: string, availableNS: string[]) {
  if (currentNS) {
    // If a namespace has been already selected, use it
    return currentNS;
  }
  // Try to get the latest namespace used
  const storedNS = getStoredNamespace(cluster);
  if (storedNS && availableNS.includes(storedNS)) {
    return storedNS;
  }
  // Try to get a namespace from the auth token
  const tokenNS = Auth.defaultNamespaceFromToken(Auth.getAuthToken() || "");
  if (tokenNS && availableNS.includes(tokenNS)) {
    // Return the default namespace in the token (if exists)
    return tokenNS;
  }
  // In other case, just return the first namespace available
  return availableNS.length ? availableNS[0] : "";
}
