// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { ThunkAction, ThunkDispatch } from "redux-thunk";
import { Kube } from "shared/Kube";
import { keyForResourceRef } from "shared/ResourceRef";
import { IK8sList, IResource, IStoreState } from "shared/types";
import { ActionType, deprecated } from "typesafe-actions";
import {
  ResourceRef as APIResourceRef,
  InstalledPackageReference,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { GetResourcesResponse } from "gen/kubeappsapis/plugins/resources/v1alpha1/resources";
import actions from "actions";
import { debounce } from "lodash";

const { createAction } = deprecated;

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

export const receiveResourceError = createAction("RECEIVE_RESOURCE_ERROR", resolve => {
  return (resource: { key: string; error: Error }) => resolve(resource);
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
    onComplete: () => void,
  ) => resolve({ pkg, refs, watch, handler, onError, onComplete });
});

export const receiveResourcesError = createAction("RECEIVE_RESOURCES_ERROR", resolve => {
  return (err: Error) => resolve(err);
});

export const closeRequestResources = createAction("CLOSE_REQUEST_RESOURCES", resolve => {
  return (pkg: InstalledPackageReference) => resolve(pkg);
});

const allActions = [
  receiveResource,
  receiveResourceError,
  requestResources,
  receiveResourcesError,
  closeRequestResources,
  requestResourceKinds,
  receiveResourceKinds,
  receiveKindsError,
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

// getResources requests and processes the responses for the resources
// associated with an installed package.
export function getResources(
  pkg: InstalledPackageReference,
  refs: APIResourceRef[],
  watch: boolean,
): ThunkAction<void, IStoreState, null, KubeAction> {
  // After resources are processed, we want to refresh the status of the
  // installed package (since other UX components rely on the status), but
  // we don't need to do this after every resource is processed, rather doing
  // it once after a bunch of resources have been processed is enough.
  // To do this, we use the lodash debounce function so that the status is refreshed
  // 2s after the last resource is processed.
  const dispatchGetInstalledPkgStatus = (
    dispatch: ThunkDispatch<IStoreState, null, KubeAction>,
    pkg: InstalledPackageReference,
  ) => dispatch(actions.installedpackages.getInstalledPkgStatus(pkg));
  const debouncedGetInstalledPkgStatus = debounce(dispatchGetInstalledPkgStatus, 2000);
  return dispatch => {
    dispatch(
      requestResources(
        pkg,
        refs,
        watch,
        (r: GetResourcesResponse) => {
          processGetResourcesResponse(r, dispatch);
          debouncedGetInstalledPkgStatus(dispatch, pkg);
        },
        (e: any) => {
          dispatch(receiveResourcesError(e));
        },
        () => {
          // The onComplete handler should only dispatch a closeRequestResources
          // action if this call to `getResources` is for watching. If it is not
          // watching resources, the server will close the request automatically
          // (and we have no book-keeping in the redux state).
          if (watch) {
            dispatch(closeRequestResources(pkg));
          }
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
  const key = keyForResourceRef(r.resourceRef);
  const manifest = r.manifest;
  const resource: IResource = JSON.parse(manifest);
  dispatch(receiveResource({ key, resource }));
}
