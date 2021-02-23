import { ThunkAction } from "redux-thunk";
import { Kube } from "shared/Kube";
import { ActionType, deprecated } from "typesafe-actions";
import ResourceRef from "../shared/ResourceRef";
import { IK8sList, IResource, IStoreState } from "../shared/types";

const { createAction } = deprecated;

export const requestResource = createAction("REQUEST_RESOURCE", resolve => {
  return (resourceID: string) => resolve(resourceID);
});

export const receiveResource = createAction("RECEIVE_RESOURCE", resolve => {
  return (resource: { key: string; resource: IResource | IK8sList<IResource, {}> }) =>
    resolve(resource);
});

export const requestResourceKinds = createAction("REQUEST_RESOURCE_KINDS");

export const receiveResourceKinds = createAction("RECEIVE_RESOURCE_KINDS", resolve => {
  return (kinds: {}) => resolve(kinds);
});

export const receiveKindsError = createAction("RECEIVE_KIND_API_ERROR", resolve => {
  return (err: Error) => resolve(err);
});

export const receiveResourceFromList = createAction("RECEIVE_RESOURCE_FROM_LIST", resolve => {
  return (resource: { key: string; resource: IResource }) => resolve(resource);
});

export const receiveResourceError = createAction("RECEIVE_RESOURCE_ERROR", resolve => {
  return (resource: { key: string; error: Error }) => resolve(resource);
});

// Takes a ResourceRef to open a WebSocket for and a handler to process messages
// from the socket.
export const openWatchResource = createAction("OPEN_WATCH_RESOURCE", resolve => {
  return (ref: ResourceRef, handler: (e: MessageEvent) => void, onError: (e: Event) => void) =>
    resolve({ ref, handler, onError });
});

export const closeWatchResource = createAction("CLOSE_WATCH_RESOURCE", resolve => {
  return (ref: ResourceRef) => resolve(ref);
});

export const addTimer = createAction("ADD_TIMER", resolve => {
  return (id: string, timer: () => void) => resolve({ id, timer });
});

export const removeTimer = createAction("REMOVE_TIMER", resolve => {
  return (id: string) => resolve(id);
});

const allActions = [
  requestResource,
  receiveResource,
  receiveResourceError,
  openWatchResource,
  closeWatchResource,
  receiveResourceFromList,
  requestResourceKinds,
  receiveResourceKinds,
  receiveKindsError,
  addTimer,
  removeTimer,
];

export type KubeAction = ActionType<typeof allActions[number]>;

export function getResourceKinds(
  cluster: string,
): ThunkAction<Promise<void>, IStoreState, null, KubeAction> {
  return async dispatch => {
    dispatch(requestResourceKinds());
    try {
      const groups = await Kube.getAPIGroups(cluster);
      const kinds = await Kube.getResourceKinds(cluster, groups);
      dispatch(receiveResourceKinds(kinds));
    } catch (e) {
      dispatch(receiveKindsError(e));
    }
  };
}

export function getResource(
  ref: ResourceRef,
  polling?: boolean,
): ThunkAction<Promise<void>, IStoreState, null, KubeAction> {
  return async (dispatch, getState) => {
    const key = ref.getResourceURL();

    // Multiple requests for a resource may happen at the same time (e.g. when
    // loading different sections of a view). This guard attempts to prevent
    // multiple requests for a resource that is already being fetched. Since
    // this action is asynchronous, it's possible that multiple requests are let
    // through, but this is not a huge concern.
    const existing = getState().kube.items[key];
    if (existing && existing.isFetching) {
      return;
    }

    // If it's not the first request, we can skip the request REDUX event
    // to avoid the loading animation
    if (!polling) {
      dispatch(requestResource(key));
    }
    try {
      const r = await ref.getResource();
      dispatch(receiveResource({ key, resource: r as IResource }));
    } catch (e) {
      dispatch(receiveResourceError({ key, error: e }));
    }
  };
}

export function getAndWatchResource(
  ref: ResourceRef,
): ThunkAction<void, IStoreState, null, KubeAction> {
  return dispatch => {
    dispatch(getResource(ref));
    dispatch(
      openWatchResource(
        ref,
        (e: MessageEvent) => {
          // If there is a timer set (fallback mechanism), remove it because
          // we are receiving events from the websocket
          dispatch(removeTimer(ref.getResourceURL()));
          const msg = JSON.parse(e.data);
          const resource: IResource = msg.object;
          const key = ref.getResourceURL();
          if (!ref.name) {
            // if the ref doesn't have a name, it's a list
            dispatch(receiveResourceFromList({ key, resource }));
          } else {
            dispatch(receiveResource({ key, resource }));
          }
        },
        () => {
          // If the Socket fails, create an interval to re-request the resource
          // every 5 seconds. This interval needs to be closed calling closeTimer
          const timer = () => dispatch(getResource(ref, true));
          dispatch(addTimer(ref.getResourceURL(), timer));
        },
      ),
    );
  };
}
