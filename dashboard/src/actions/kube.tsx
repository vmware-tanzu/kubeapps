import { ThunkAction } from "redux-thunk";
import { ActionType, createAction } from "typesafe-actions";
import { Kube } from "../shared/Kube";
import { IKubeItem, IStoreState } from "../shared/types";

export const requestResource = createAction("REQUEST_RESOURCE", resolve => {
  return (resource: { [s: string]: IKubeItem }) => resolve(resource);
});

export const receiveResource = createAction("RECEIVE_RESOURCE", resolve => {
  return (resource: { [s: string]: IKubeItem }) => resolve(resource);
});

export const errorKube = createAction("ERROR_KUBE", resolve => {
  return (resource: { [s: string]: IKubeItem }) => resolve(resource);
});

const allActions = [requestResource, receiveResource, errorKube];

export type KubeAction = ActionType<typeof allActions[number]>;

export function getResource(
  apiVersion: string,
  resource: string,
  namespace: string,
  name: string,
  query?: string,
): ThunkAction<Promise<void>, IStoreState, null, KubeAction> {
  return async dispatch => {
    const key = `${apiVersion}/${namespace}/${resource}/${name}${query ? `/${query}` : ""}`;
    dispatch(requestResource({ [key]: { isFetching: true } }));
    try {
      const r = await Kube.getResource(apiVersion, resource, namespace, name, query);
      dispatch(receiveResource({ [key]: { isFetching: false, item: r } }));
    } catch (e) {
      dispatch(errorKube({ [key]: { isFetching: false, error: e } }));
    }
  };
}
