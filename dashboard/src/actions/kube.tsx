import { ThunkAction } from "redux-thunk";
import { ActionType, createAction } from "typesafe-actions";
import { Kube } from "../shared/Kube";
import { IResource, IStoreState } from "../shared/types";

export const requestResource = createAction("REQUEST_RESOURCE", resolve => {
  return (resourceID: string) => resolve(resourceID);
});

export const receiveResource = createAction("RECEIVE_RESOURCE", resolve => {
  return (resource: { key: string; resource: IResource }) => resolve(resource);
});

export const receiveResourceError = createAction("RECEIVE_RESOURCE_ERROR", resolve => {
  return (resource: { key: string; error: Error }) => resolve(resource);
});

const allActions = [requestResource, receiveResource, receiveResourceError];

export type KubeAction = ActionType<typeof allActions[number]>;

export function getResource(
  apiVersion: string,
  resource: string,
  namespace: string,
  name: string,
  query?: string,
): ThunkAction<Promise<void>, IStoreState, null, KubeAction> {
  return async dispatch => {
    const key = Kube.getResourceURL(apiVersion, resource, namespace, name, query);
    dispatch(requestResource(key));
    try {
      const r = await Kube.getResource(apiVersion, resource, namespace, name, query);
      dispatch(receiveResource({ key, resource: r }));
    } catch (e) {
      dispatch(receiveResourceError({ key, error: e }));
    }
  };
}
