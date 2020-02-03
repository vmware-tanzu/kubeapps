import { ThunkAction } from "redux-thunk";

import { ActionType, createAction } from "typesafe-actions";

import Namespace from "../shared/Namespace";
import { IResource, IStoreState } from "../shared/types";

export const requestNamespace = createAction("REQUEST_NAMESPACE", resolve => {
  return (namespace: string) => resolve(namespace);
});
export const receiveNamespace = createAction("RECEIVE_NAMESPACE", resolve => {
  return (namespace: IResource) => resolve(namespace);
});

export const setNamespace = createAction("SET_NAMESPACE", resolve => {
  return (namespace: string) => resolve(namespace);
});

export const postNamespace = createAction("CREATE_NAMESPACE", resolve => {
  return (namespace: string) => resolve(namespace);
});

export const receiveNamespaces = createAction("RECEIVE_NAMESPACES", resolve => {
  return (namespaces: string[]) => resolve(namespaces);
});

export const errorNamespaces = createAction("ERROR_NAMESPACES", resolve => {
  return (err: Error, op: string) => resolve({ err, op });
});

export const clearNamespaces = createAction("CLEAR_NAMESPACES");

const allActions = [
  requestNamespace,
  receiveNamespace,
  setNamespace,
  receiveNamespaces,
  errorNamespaces,
  clearNamespaces,
  postNamespace,
];
export type NamespaceAction = ActionType<typeof allActions[number]>;

export function fetchNamespaces(): ThunkAction<Promise<void>, IStoreState, null, NamespaceAction> {
  return async dispatch => {
    try {
      const namespaces = await Namespace.list();
      const namespaceStrings = namespaces.items.map((n: IResource) => n.metadata.name);
      dispatch(receiveNamespaces(namespaceStrings));
    } catch (e) {
      dispatch(errorNamespaces(e, "list"));
      return;
    }
  };
}

export function createNamespace(
  ns: string,
): ThunkAction<Promise<boolean>, IStoreState, null, NamespaceAction> {
  return async dispatch => {
    try {
      await Namespace.create(ns);
      dispatch(postNamespace(ns));
      dispatch(fetchNamespaces());
      return true;
    } catch (e) {
      dispatch(errorNamespaces(e, "create"));
      return false;
    }
  };
}

export function getNamespace(
  ns: string,
): ThunkAction<Promise<boolean>, IStoreState, null, NamespaceAction> {
  return async dispatch => {
    try {
      dispatch(requestNamespace(ns));
      const namespace = await Namespace.get(ns);
      dispatch(receiveNamespace(namespace));
      return true;
    } catch (e) {
      dispatch(errorNamespaces(e, "get"));
      return false;
    }
  };
}
