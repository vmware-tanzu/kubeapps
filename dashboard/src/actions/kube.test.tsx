import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";
import { getType } from "typesafe-actions";

import actions from ".";
import { Kube } from "../shared/Kube";
import ResourceRef, { fromCRD } from "../shared/ResourceRef";
import { IClusterServiceVersionCRD, IKubeState, IResource } from "../shared/types";

const mockStore = configureMockStore([thunk]);
const clusterName = "cluster-name";

let store: any;

beforeEach(() => {
  store = mockStore({
    kube: {
      items: {},
      sockets: {},
    } as IKubeState,
  });
});

const getResourceOrig = Kube.getResource;
let getResourceMock: jest.Mock;
beforeEach(() => {
  getResourceMock = jest.fn(() => []);
  Kube.getResource = getResourceMock;
});
afterEach(() => {
  Kube.getResource = getResourceOrig;
});

describe("getResource", () => {
  it("fetches a resource", async () => {
    const expectedActions = [
      {
        type: getType(actions.kube.requestResource),
        payload: `api/clusters/${clusterName}/api/v1/namespaces/default/services/foo`,
      },
      {
        type: getType(actions.kube.receiveResource),
        payload: {
          key: `api/clusters/${clusterName}/api/v1/namespaces/default/services/foo`,
          resource: [],
        },
      },
    ];
    const r = {
      apiVersion: "v1",
      kind: "Service",
      metadata: {
        name: "foo",
        namespace: "default",
      },
    } as IResource;

    const ref = new ResourceRef(r, clusterName);

    await store.dispatch(actions.kube.getResource(ref));
    expect(store.getActions()).toEqual(expectedActions);
    expect(getResourceMock).toHaveBeenCalledWith(clusterName, "v1", "services", "default", "foo");
  });

  it("does not fetch a resource that is already being fetched", async () => {
    store = mockStore({
      kube: {
        items: {
          [`api/clusters/${clusterName}/api/v1/namespaces/default/services/foo`]: {
            isFetching: true,
          },
        },
      },
    });

    const r = {
      apiVersion: "v1",
      kind: "Service",
      metadata: {
        name: "foo",
        namespace: "default",
      },
    } as IResource;

    const ref = new ResourceRef(r, clusterName);

    await store.dispatch(actions.kube.getResource(ref));
    expect(store.getActions()).toEqual([]);
    expect(getResourceMock).not.toBeCalled();
  });
});

describe("getAndWatchResource", () => {
  it("dispatches a getResource and openWatchResource action", () => {
    const r = {
      apiVersion: "v1",
      kind: "Service",
      metadata: {
        name: "foo",
        namespace: "default",
      },
    } as IResource;

    const ref = new ResourceRef(r, clusterName);

    const expectedActions = [
      {
        type: getType(actions.kube.requestResource),
        payload: `api/clusters/${clusterName}/api/v1/namespaces/default/services/foo`,
      },
      {
        type: getType(actions.kube.openWatchResource),
        payload: {
          ref,
          handler: expect.any(Function),
          onError: { onErrorHandler: expect.any(Function), closeTimer: expect.any(Function) },
        },
      },
    ];

    store.dispatch(actions.kube.getAndWatchResource(ref));
    expect(store.getActions()).toEqual(expectedActions);
    expect(getResourceMock).toHaveBeenCalledWith(clusterName, "v1", "services", "default", "foo");
  });

  it("dispatches a getResource and openWatchResource action for a list", () => {
    const ref = {
      kind: "Service",
      name: "",
      version: "v1",
    } as IClusterServiceVersionCRD;
    const svc = {
      apiVersion: "v1",
      kind: "Service",
      metadata: {
        name: "foo",
        namespace: "default",
      },
    } as IResource;

    const r = fromCRD(ref, clusterName, "default", {});
    const expectedActions = [
      {
        type: getType(actions.kube.requestResource),
        payload: `api/clusters/${clusterName}/api/v1/namespaces/default/services`,
      },
      {
        type: getType(actions.kube.openWatchResource),
        payload: {
          ref: r,
          handler: expect.any(Function),
          onError: { onErrorHandler: expect.any(Function), closeTimer: expect.any(Function) },
        },
      },
    ];

    store.dispatch(actions.kube.getAndWatchResource(r));
    const testActions = store.getActions();
    expect(testActions).toEqual(expectedActions);
    expect(getResourceMock).toHaveBeenCalledWith(
      clusterName,
      "v1",
      "services",
      "default",
      undefined,
    );

    const watchFunction = (testActions[1].payload as any).handler as (e: any) => void;
    watchFunction({ data: `{"object": ${JSON.stringify(svc)}}` });
    const newAction = {
      type: getType(actions.kube.receiveResourceFromList),
      payload: {
        key: `api/clusters/${clusterName}/api/v1/namespaces/default/services`,
        resource: svc,
      },
    };
    const expectedUpdatedActions = expectedActions.concat(newAction as any);
    const updatedActions = store.getActions();
    expect(updatedActions).toEqual(expectedUpdatedActions);
  });
});
