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
      return { ...state, items: { ...state.items, ...action.payload } };
    case getType(actions.kube.receiveResource):
      return { ...state, items: { ...state.items, ...action.payload } };
    case getType(actions.kube.receiveResourceError):
      return { ...state, items: { ...state.items, ...action.payload } };
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
