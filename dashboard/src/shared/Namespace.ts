import { get } from "lodash";
import { Auth } from "./Auth";
import { axiosWithAuth } from "./AxiosInstance";
import { ForbiddenError, IResource, NotFoundError } from "./types";
import * as url from "./url";
import { KubeappsGrpcClient } from "./KubeappsGrpcClient";

export default class Namespace {
  private static resourcesClient = () => new KubeappsGrpcClient().getResourcesServiceClientImpl();

  public static async list(cluster: string) {
    // This call is hitting an actual backend endpoint (see pkg/http-handler.go)
    // while the other two calls (create, get) have been updated to use the
    // resources client rather than the k8s API server.
    const { data } = await axiosWithAuth.get<{ namespaces: IResource[] }>(
      url.backend.namespaces.list(cluster),
    );
    return data;
  }

  public static async create(cluster: string, namespace: string) {
    await this.resourcesClient().CreateNamespace({
      context: {
        cluster,
        namespace,
      },
    });
  }

  public static async exists(cluster: string, namespace: string) {
    try {
      const { exists } = await this.resourcesClient().CheckNamespaceExists({
        context: {
          cluster,
          namespace,
        },
      });

      return exists;
    } catch (e: any) {
      switch (e.constructor) {
        case ForbiddenError:
          throw new ForbiddenError(
            `You don't have sufficient permissions to use the namespace ${namespace}`,
          );
        case NotFoundError:
          throw new NotFoundError(`Namespace ${namespace} not found. Create it before using it.`);
        default:
          throw e;
      }
    }
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
