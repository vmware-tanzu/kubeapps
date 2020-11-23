import { get } from "lodash";
import { Auth } from "./Auth";
import { axiosWithAuth } from "./AxiosInstance";
import { ForbiddenError, IResource, NotFoundError } from "./types";
import * as url from "./url";

export default class Namespace {
  public static async list(cluster: string) {
    // This call is hitting an actual backend endpoint (see pkg/http-handler.go)
    // while the other calls (create, get) are hitting the k8s API via the
    // frountend nginx.
    const { data } = await axiosWithAuth.get<{ namespaces: IResource[] }>(
      url.backend.namespaces.list(cluster),
    );
    return data;
  }

  public static async create(cluster: string, namespace: string) {
    const { data } = await axiosWithAuth.post<IResource>(url.api.k8s.namespaces(cluster), {
      apiVersion: "v1",
      kind: "Namespace",
      metadata: {
        name: namespace,
      },
    });
    return data;
  }

  public static async get(cluster: string, namespace: string) {
    try {
      const { data } = await axiosWithAuth.get<IResource>(
        url.api.k8s.namespace(cluster, namespace),
      );
      return data;
    } catch (err) {
      switch (err.constructor) {
        case ForbiddenError:
          throw new ForbiddenError(
            `You don't have sufficient permissions to use the namespace ${namespace}`,
          );
        case NotFoundError:
          throw new NotFoundError(`Namespace ${namespace} not found. Create it before using it.`);
        default:
          throw err;
      }
    }
  }
}

// Set of namespaces used accross the applications as default and "all ns" placeholders
// export const definedNamespaces = {
//   all: "_all",
// };

// The namespace information will contain a map[cluster]:namespace with the default namespaces
const namespaceKey = "kubeapps_namespace";

function parseStoredNS() {
  const ns = localStorage.getItem(namespaceKey) || "{}";
  let parsedNS = {};
  try {
    parsedNS = JSON.parse(ns);
  } catch (e) {
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
