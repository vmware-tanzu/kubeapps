import { ThunkAction } from "redux-thunk";

import { ActionType, createAction } from "typesafe-actions";

import Namespace from "../shared/Namespace";
import { IResource, IStoreState } from "../shared/types";

export const setNamespace = createAction("SET_NAMESPACE", resolve => {
  return (namespace: string) => resolve(namespace);
});

export const receiveNamespaces = createAction("RECEIVE_NAMESPACES", resolve => {
  return (namespaces: string[]) => resolve(namespaces);
});

const allActions = [setNamespace, receiveNamespaces];
export type NamespaceAction = ActionType<typeof allActions[number]>;

export function fetchNamespaces(): ThunkAction<Promise<void>, IStoreState, null, NamespaceAction> {
  return async dispatch => {
    try {
      const namespaces = await Namespace.list();
      const namespaceStrings = namespaces.items.map((n: IResource) => n.metadata.name);
      dispatch(receiveNamespaces(namespaceStrings));
    } catch (e) {
      // TODO: handle namespace call error
      return;
    }
  };
}
