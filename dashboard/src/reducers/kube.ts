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
      const rKey = Object.keys(action.payload)[0];
      const rResource = action.payload[rKey];
      const receivedItem = { [rKey]: { isFetching: false, item: rResource } };
      return { ...state, items: { ...state.items, ...receivedItem } };
    case getType(actions.kube.receiveResourceError):
      const eKey = Object.keys(action.payload)[0];
      const eResource = action.payload[eKey];
      const erroredItem = { [eKey]: { isFetching: false, error: eResource } };
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
