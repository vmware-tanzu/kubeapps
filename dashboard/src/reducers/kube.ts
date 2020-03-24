import { LOCATION_CHANGE, LocationChangeAction } from "connected-react-router";

import { getType } from "typesafe-actions";
import actions from "../actions";
import { KubeAction } from "../actions/kube";
import { IK8sList, IKubeState, IResource } from "../shared/types";

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
    case getType(actions.kube.receiveResourceFromList):
      const stateListItem = state.items[action.payload.key].item as IK8sList<IResource, {}>;
      const newItem = action.payload.resource as IResource;
      if (!stateListItem || !stateListItem.items) {
        return {
          ...state,
          items: {
            ...state.items,
            [action.payload.key]: {
              isFetching: false,
              item: { ...stateListItem, items: [newItem] },
            },
          },
        };
      }
      const updatedItems = stateListItem.items.map(it => {
        if (it.metadata.selfLink === newItem.metadata.selfLink) {
          return action.payload.resource as IResource;
        }
        return it;
      });
      return {
        ...state,
        items: {
          ...state.items,
          [action.payload.key]: {
            isFetching: false,
            item: { ...stateListItem, items: updatedItems },
          },
        },
      };
    case getType(actions.kube.receiveResourceError):
      const erroredItem = {
        [action.payload.key]: { isFetching: false, error: action.payload.error },
      };
      return { ...state, items: { ...state.items, ...erroredItem } };
    case getType(actions.kube.openWatchResource):
      const { ref, handler, onError } = action.payload;
      key = ref.watchResourceURL();
      if (state.sockets[key]) {
        // Socket for this resource already open, do nothing
        return state;
      }
      const socket = ref.watchResource();
      socket.addEventListener("message", handler);
      const { onErrorHandler, closeTimer } = onError;
      socket.addEventListener("error", onErrorHandler);
      return {
        ...state,
        sockets: {
          ...state.sockets,
          [key]: { socket, closeTimer },
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
        foundSocket.socket.close();
        foundSocket.closeTimer();
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
