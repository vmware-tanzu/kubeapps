import { LOCATION_CHANGE, LocationChangeAction } from "connected-react-router";
import * as _ from "lodash";

import { getType } from "typesafe-actions";
import actions from "../actions";
import { KubeAction } from "../actions/kube";
import { IKubeState } from "../shared/types";

const initialState: IKubeState = {
  items: {},
};

const kubeReducer = (
  state: IKubeState = initialState,
  action: KubeAction | LocationChangeAction,
): IKubeState => {
  switch (action.type) {
    case getType(actions.kube.requestResource):
      const requestedItem = { [action.payload]: { isFetching: true } };
      return { ...state, items: { ...state.items, ...requestedItem } };
    case getType(actions.kube.receiveResource):
      const receivedItem = {
        [action.payload.key]: { isFetching: false, item: action.payload.resource },
      };
      return { ...state, items: { ...state.items, ...receivedItem } };
    case getType(actions.kube.receiveResourceError):
      const erroredItem = {
        [action.payload.key]: { isFetching: false, error: action.payload.error },
      };
      return { ...state, items: { ...state.items, ...erroredItem } };
    case LOCATION_CHANGE:
      return {
        ...state,
        items: {},
      };
    default:
  }
  return state;
};

export default kubeReducer;
