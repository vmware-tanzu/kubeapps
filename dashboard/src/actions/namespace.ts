import { Action, Dispatch } from "redux";
import { ThunkAction } from "redux-thunk";

import { ActionType, createActionDeprecated } from "typesafe-actions";

import Namespace from "../shared/Namespace";
import { IResource, IStoreState } from "../shared/types";

export const setNamespace = createActionDeprecated("SET_NAMESPACE", (namespace: string) => {
  return {
    namespace,
    type: "SET_NAMESPACE",
  };
});

export const receiveNamespaces = createActionDeprecated(
  "RECEIVE_NAMESPACES",
  (namespaces: string[]) => {
    return {
      namespaces,
      type: "RECEIVE_NAMESPACES",
    };
  },
);

const allActions = [setNamespace, receiveNamespaces];
export type NamespaceAction = ActionType<typeof allActions[number]>;

export function fetchNamespaces(): ThunkAction<Promise<void>, IStoreState, void, Action> {
  return async (dispatch: Dispatch): Promise<void> => {
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
