import { LOCATION_CHANGE, LocationChangeAction } from "react-router-redux";
import { getType } from "typesafe-actions";

import actions from "../actions";
import { NamespaceAction } from "../actions/namespace";

export interface INamespaceState {
  current: string;
  namespaces: string[];
}

const initialState: INamespaceState = {
  current: localStorage.getItem("kubeapps_namespace") || "default",
  namespaces: ["default"],
};

const namespaceReducer = (
  state: INamespaceState = initialState,
  action: NamespaceAction | LocationChangeAction,
): INamespaceState => {
  switch (action.type) {
    case getType(actions.namespace.receiveNamespaces):
      return { ...state, namespaces: action.namespaces };
    case getType(actions.namespace.setNamespace):
      return { ...state, current: action.namespace };
    case LOCATION_CHANGE:
      const pathname = action.payload.pathname;
      // looks for /ns/:namespace in URL
      const matches = pathname.match(/\/ns\/([^/]*)/);
      if (matches) {
        return { ...state, current: matches[1] };
      }
    default:
  }
  return state;
};

export default namespaceReducer;
