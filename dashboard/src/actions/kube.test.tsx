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

    const ref = new ResourceRef(r, clusterName, "services", true);

    await store.dispatch(actions.kube.getResource(ref));
    expect(store.getActions()).toEqual(expectedActions);
    expect(getResourceMock).toHaveBeenCalledWith(
      clusterName,
      "v1",
      "services",
      true,
      "default",
      "foo",
    );
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

    const ref = new ResourceRef(r, clusterName, "services", true);

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

    const ref = new ResourceRef(r, clusterName, "services", true);

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
          onError: expect.any(Function),
        },
      },
    ];

    store.dispatch(actions.kube.getAndWatchResource(ref));
    expect(store.getActions()).toEqual(expectedActions);
    expect(getResourceMock).toHaveBeenCalledWith(
      clusterName,
      "v1",
      "services",
      true,
      "default",
      "foo",
    );
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

    const r = fromCRD(
      ref,
      { apiVersion: "v1", plural: "services", namespaced: true },
      clusterName,
      "default",
      {},
    );
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
          onError: expect.any(Function),
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
      true,
      "default",
      undefined,
    );

    const watchFunction = (testActions[1].payload as any).handler as (e: any) => void;
    watchFunction({ data: `{"object": ${JSON.stringify(svc)}}` });
    const newActions = [
      {
        type: getType(actions.kube.removeTimer),
        payload: `api/clusters/${clusterName}/api/v1/namespaces/default/services`,
      },
      {
        type: getType(actions.kube.receiveResourceFromList),
        payload: {
          key: `api/clusters/${clusterName}/api/v1/namespaces/default/services`,
          resource: svc,
        },
      },
    ];
    const expectedUpdatedActions = expectedActions.concat(newActions as any);
    const updatedActions = store.getActions();
    expect(updatedActions).toEqual(expectedUpdatedActions);
  });

  it("adds a timer and removes it if a new event is received", () => {
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

    const r = fromCRD(
      ref,
      { apiVersion: "v1", plural: "services", namespaced: true },
      clusterName,
      "default",
      {},
    );
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
          onError: expect.any(Function),
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
      true,
      "default",
      undefined,
    );

    const watchFunction = (testActions[1].payload as any).handler as (e: any) => void;
    const onErrorFunction = (testActions[1].payload as any).onError as () => void;
    onErrorFunction();
    watchFunction({ data: `{"object": ${JSON.stringify(svc)}}` });
    const newActions = [
      {
        type: getType(actions.kube.addTimer),
        payload: {
          id: `api/clusters/${clusterName}/api/v1/namespaces/default/services`,
          timer: expect.any(Function),
        },
      },
      {
        type: getType(actions.kube.removeTimer),
        payload: `api/clusters/${clusterName}/api/v1/namespaces/default/services`,
      },
      {
        type: getType(actions.kube.receiveResourceFromList),
        payload: {
          key: `api/clusters/${clusterName}/api/v1/namespaces/default/services`,
          resource: svc,
        },
      },
    ];
    const expectedUpdatedActions = expectedActions.concat(newActions as any);
    const updatedActions = store.getActions();
    expect(updatedActions).toEqual(expectedUpdatedActions);
  });
});

describe("getResourceKinds", () => {
  beforeEach(() => {
    Kube.getAPIGroups = jest.fn().mockResolvedValue([]);
    Kube.getResourceKinds = jest.fn().mockResolvedValue({});
  });
  afterEach(() => {
    jest.resetAllMocks();
  });

  it("retrieves resource kinds", async () => {
    const groups = ["foo"];
    const kinds = { test: "data" };
    Kube.getAPIGroups = jest.fn().mockResolvedValue(groups);
    Kube.getResourceKinds = jest.fn().mockResolvedValue(kinds);
    const expectedActions = [
      {
        type: getType(actions.kube.requestResourceKinds),
      },
      {
        type: getType(actions.kube.receiveResourceKinds),
        payload: kinds,
      },
    ];
    await store.dispatch(actions.kube.getResourceKinds("cluster-1"));

    const testActions = store.getActions();
    expect(testActions).toEqual(expectedActions);
    expect(Kube.getAPIGroups).toHaveBeenCalledWith("cluster-1");
    expect(Kube.getResourceKinds).toHaveBeenCalledWith("cluster-1", groups);
  });

  it("returns an error if fails to retrieve kinds", async () => {
    const groups = ["foo"];
    Kube.getAPIGroups = jest.fn().mockResolvedValue(groups);
    Kube.getResourceKinds = jest.fn().mockRejectedValue(new Error("boom!"));
    const expectedActions = [
      {
        type: getType(actions.kube.requestResourceKinds),
      },
      {
        type: getType(actions.kube.receiveKindsError),
        payload: new Error("boom!"),
      },
    ];
    await store.dispatch(actions.kube.getResourceKinds("cluster-1"));

    const testActions = store.getActions();
    expect(testActions).toEqual(expectedActions);
    expect(Kube.getAPIGroups).toHaveBeenCalledWith("cluster-1");
    expect(Kube.getResourceKinds).toHaveBeenCalledWith("cluster-1", groups);
  });
});
