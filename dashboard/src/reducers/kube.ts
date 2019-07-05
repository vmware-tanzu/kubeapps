import { LOCATION_CHANGE, LocationChangeAction } from "connected-react-router";

import { getType } from "typesafe-actions";
import actions from "../actions";
import { KubeAction } from "../actions/kube";
import { IKubeState } from "../shared/types";

const initialState: IKubeState = {
  items: {},
  sockets: {},
};

const kubeReducer = (
  state: IKubeState = initialState,
  action: KubeAction | LocationChangeAction,
): IKubeState => {
  let key: string;
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
    case getType(actions.kube.openWatchResource):
      const { ref, handler } = action.payload;
      key = ref.watchResourceURL();
      if (state.sockets[key]) {
        // Socket for this resource already open, do nothing
        return state;
      }
      const socket = ref.watchResource();
      socket.addEventListener("message", handler);
      return {
        ...state,
        sockets: {
          ...state.sockets,
          [key]: socket,
        },
      };
    // TODO(adnan): this won't handle cases where one component closes a socket
    // another one is using. Whilst not a problem today, a reference counter
    // approach could be used here to enable this in the future.
    case getType(actions.kube.closeWatchResource):
      key = action.payload.watchResourceURL();
      const { sockets } = state;
      const { [key]: foundSocket, ...otherSockets } = sockets;
      // close the socket if it exists
      if (foundSocket !== undefined) {
        foundSocket.close();
      }
      return {
        ...state,
        sockets: otherSockets,
      };
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
