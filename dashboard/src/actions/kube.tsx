import { ThunkAction, ThunkDispatch } from "redux-thunk";
import { Kube } from "shared/Kube";
import ResourceRef, { keyForResourceRef } from "shared/ResourceRef";
import { IK8sList, IResource, IStoreState } from "shared/types";
import { ActionType, deprecated } from "typesafe-actions";
import {
  ResourceRef as APIResourceRef,
  InstalledPackageReference,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { GetResourcesResponse } from "gen/kubeappsapis/plugins/resources/v1alpha1/resources";

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

// requestResources takes a ResourceRef[] and subscribes to an observable to
// process the responses as they arrive.
export const requestResources = createAction("REQUEST_RESOURCES", resolve => {
  return (
    pkg: InstalledPackageReference,
    refs: APIResourceRef[],
    watch: boolean,
    handler: (r: GetResourcesResponse) => void,
    onError: (e: Event) => void,
    onComplete: (pkg: InstalledPackageReference) => void,
  ) => resolve({ pkg, refs, watch, handler, onError, onComplete });
});

export const receiveResourcesError = createAction("RECEIVE_RESOURCES_ERROR", resolve => {
  return (err: Error) => resolve(err);
});

export const closeRequestResources = createAction("CLOSE_REQUEST_RESOURCES", resolve => {
  return (pkg: InstalledPackageReference) => resolve(pkg);
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
  requestResources,
  receiveResourcesError,
  closeRequestResources,
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
    } catch (e: any) {
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
    } catch (e: any) {
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

// getResources requests and processes the responses for the resources
// associated with an installed package.
export function getResources(
  pkg: InstalledPackageReference,
  refs: APIResourceRef[],
  watch: boolean,
): ThunkAction<void, IStoreState, null, KubeAction> {
  return dispatch => {
    dispatch(
      requestResources(
        pkg,
        refs,
        watch,
        (r: GetResourcesResponse) => {
          processGetResourcesResponse(r, dispatch);
        },
        (e: any) => {
          dispatch(receiveResourcesError(e));
        },
        (pkg: InstalledPackageReference) => {
          dispatch(closeRequestResources(pkg));
        },
      ),
    );
  };
}

export function processGetResourcesResponse(
  r: GetResourcesResponse,
  dispatch: ThunkDispatch<IStoreState, null, KubeAction>,
) {
  if (!r.resourceRef) {
    dispatch(receiveResourcesError(new Error("received resource without a resource reference")));
    return;
  }
  const key = keyForResourceRef(
    r.resourceRef!.apiVersion,
    r.resourceRef!.kind,
    r.resourceRef!.namespace,
    r.resourceRef!.name,
  );
  const manifest = new TextDecoder().decode(r.manifest!.value);
  const resource: IResource = JSON.parse(manifest);
  dispatch(receiveResource({ key, resource }));
}
