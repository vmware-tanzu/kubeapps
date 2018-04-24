import { Dispatch } from "react-redux";
import { createAction, getReturnOfExpression } from "typesafe-actions";

import Namespace from "../shared/Namespace";
import { IResource, IStoreState } from "../shared/types";

export const setNamespace = createAction("SET_NAMESPACE", (namespace: string) => {
  return {
    namespace,
    type: "SET_NAMESPACE",
  };
});

export const receiveNamespaces = createAction("RECEIVE_NAMESPACES", (namespaces: string[]) => {
  return {
    namespaces,
    type: "RECEIVE_NAMESPACES",
  };
});

const allActions = [setNamespace, receiveNamespaces].map(getReturnOfExpression);
export type NamespaceAction = typeof allActions[number];

export function fetchNamespaces() {
  return async (dispatch: Dispatch<IStoreState>) => {
    try {
      const namespaces = await Namespace.list();
      const namespaceStrings = namespaces.items.map((n: IResource) => n.metadata.name);
      return dispatch(receiveNamespaces(namespaceStrings));
    } catch (e) {
      // TODO: handle namespace call error
      return;
    }
  };
}
