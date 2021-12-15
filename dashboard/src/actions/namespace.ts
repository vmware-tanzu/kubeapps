import { get } from "lodash";
import { ThunkAction } from "redux-thunk";
import { Kube } from "shared/Kube";
import Namespace, { setStoredNamespace } from "shared/Namespace";
import { IResource, IStoreState } from "shared/types";
import { ActionType, deprecated } from "typesafe-actions";

const { createAction } = deprecated;

export const requestNamespaceExists = createAction("REQUEST_NAMESPACE", resolve => {
  return (cluster: string, namespace: string) => resolve({ cluster, namespace });
});
export const receiveNamespaceExists = createAction("RECEIVE_NAMESPACE", resolve => {
  return (cluster: string, namespace: string) => resolve({ cluster, namespace });
});

export const setNamespaceState = createAction("SET_NAMESPACE", resolve => {
  return (cluster: string, namespace: string) => resolve({ cluster, namespace });
});

export const postNamespace = createAction("CREATE_NAMESPACE", resolve => {
  return (cluster: string, namespace: string) => resolve({ cluster, namespace });
});

export const receiveNamespaces = createAction("RECEIVE_NAMESPACES", resolve => {
  return (cluster: string, namespaces: string[]) => resolve({ cluster, namespaces });
});

export const errorNamespaces = createAction("ERROR_NAMESPACES", resolve => {
  return (cluster: string, err: Error, op: string) => resolve({ cluster, err, op });
});

export const setAllowCreate = createAction("ALLOW_CREATE_NAMESPACE", resolve => {
  return (cluster: string, allowed: boolean) => resolve({ cluster, allowed });
});

export const clearClusters = createAction("CLEAR_CLUSTERS");

const allActions = [
  requestNamespaceExists,
  receiveNamespaceExists,
  setNamespaceState,
  receiveNamespaces,
  errorNamespaces,
  clearClusters,
  postNamespace,
  setAllowCreate,
];
export type NamespaceAction = ActionType<typeof allActions[number]>;

export function fetchNamespaces(
  cluster: string,
): ThunkAction<Promise<string[]>, IStoreState, null, NamespaceAction> {
  return async dispatch => {
    try {
      const namespaceList = await Namespace.list(cluster);
      const namespaceStrings = get(namespaceList, "namespaces", []).map(
        (n: IResource) => n.metadata.name,
      );
      if (namespaceStrings.length === 0) {
        dispatch(
          errorNamespaces(
            cluster,
            new Error("The current account does not have access to any namespaces"),
            "list",
          ),
        );
        return [];
      }
      dispatch(receiveNamespaces(cluster, namespaceStrings));
      return namespaceStrings;
    } catch (e: any) {
      dispatch(errorNamespaces(cluster, e, "list"));
      return [];
    }
  };
}

export function createNamespace(
  cluster: string,
  ns: string,
): ThunkAction<Promise<boolean>, IStoreState, null, NamespaceAction> {
  return async dispatch => {
    try {
      await Namespace.create(cluster, ns);
      dispatch(postNamespace(cluster, ns));
      dispatch(fetchNamespaces(cluster));
      return true;
    } catch (e: any) {
      dispatch(errorNamespaces(cluster, e, "create"));
      return false;
    }
  };
}

export function checkNamespaceExists(
  cluster: string,
  ns: string,
): ThunkAction<Promise<boolean>, IStoreState, null, NamespaceAction> {
  return async dispatch => {
    try {
      dispatch(requestNamespaceExists(cluster, ns));
      const exists = await Namespace.exists(cluster, ns);
      if (exists) {
        dispatch(receiveNamespaceExists(cluster, ns));
      }
      return exists;
    } catch (e: any) {
      dispatch(errorNamespaces(cluster, e, "get"));
      return false;
    }
  };
}

export function setNamespace(
  cluster: string,
  ns: string,
): ThunkAction<Promise<void>, IStoreState, null, NamespaceAction> {
  return async dispatch => {
    setStoredNamespace(cluster, ns);
    dispatch(setNamespaceState(cluster, ns));
  };
}

export function canCreate(
  cluster: string,
): ThunkAction<Promise<void>, IStoreState, null, NamespaceAction> {
  return async dispatch => {
    try {
      const allowed = await Kube.canI(cluster, "", "namespaces", "create", "");
      dispatch(setAllowCreate(cluster, allowed));
    } catch (e: any) {
      dispatch(errorNamespaces(cluster, e, "get"));
    }
  };
}
