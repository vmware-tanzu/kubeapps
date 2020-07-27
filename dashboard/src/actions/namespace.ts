import { ThunkAction } from "redux-thunk";

import { ActionType, createAction } from "typesafe-actions";

import Namespace from "../shared/Namespace";
import { IResource, IStoreState } from "../shared/types";

export const requestNamespace = createAction("REQUEST_NAMESPACE", resolve => {
  return (cluster: string, namespace: string) => resolve({ cluster, namespace });
});
export const receiveNamespace = createAction("RECEIVE_NAMESPACE", resolve => {
  return (cluster: string, namespace: IResource) => resolve({ cluster, namespace });
});

export const setNamespace = createAction("SET_NAMESPACE", resolve => {
  return (namespace: string) => resolve(namespace);
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

export const clearClusters = createAction("CLEAR_CLUSTERS");

const allActions = [
  requestNamespace,
  receiveNamespace,
  setNamespace,
  receiveNamespaces,
  errorNamespaces,
  clearClusters,
  postNamespace,
];
export type NamespaceAction = ActionType<typeof allActions[number]>;

export function fetchNamespaces(
  cluster: string,
): ThunkAction<Promise<void>, IStoreState, null, NamespaceAction> {
  return async dispatch => {
    try {
      const namespaceList = await Namespace.list(cluster);
      const namespaceStrings = namespaceList.namespaces.map((n: IResource) => n.metadata.name);
      dispatch(receiveNamespaces(cluster, namespaceStrings));
    } catch (e) {
      dispatch(errorNamespaces(cluster, e, "list"));
      return;
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
    } catch (e) {
      dispatch(errorNamespaces(cluster, e, "create"));
      return false;
    }
  };
}

export function getNamespace(
  cluster: string,
  ns: string,
): ThunkAction<Promise<boolean>, IStoreState, null, NamespaceAction> {
  return async dispatch => {
    try {
      dispatch(requestNamespace(cluster, ns));
      const namespace = await Namespace.get(cluster, ns);
      dispatch(receiveNamespace(cluster, namespace));
      return true;
    } catch (e) {
      dispatch(errorNamespaces(cluster, e, "get"));
      return false;
    }
  };
}
