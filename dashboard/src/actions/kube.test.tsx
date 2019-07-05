import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";
import { getType } from "typesafe-actions";

import actions from ".";
import { Kube } from "../shared/Kube";
import ResourceRef from "../shared/ResourceRef";
import { IKubeState, IResource } from "../shared/types";

const mockStore = configureMockStore([thunk]);

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
        payload: "api/kube/api/v1/namespaces/default/services/foo",
      },
      {
        type: getType(actions.kube.receiveResource),
        payload: {
          key: "api/kube/api/v1/namespaces/default/services/foo",
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

    const ref = new ResourceRef(r);

    await store.dispatch(actions.kube.getResource(ref));
    expect(store.getActions()).toEqual(expectedActions);
    expect(getResourceMock).toHaveBeenCalledWith("v1", "services", "default", "foo");
  });

  it("does not fetch a resource that is already being fetched", async () => {
    store = mockStore({
      kube: {
        items: {
          "api/kube/api/v1/namespaces/default/services/foo": { isFetching: true },
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

    const ref = new ResourceRef(r);

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

    const ref = new ResourceRef(r);

    const expectedActions = [
      {
        type: getType(actions.kube.requestResource),
        payload: "api/kube/api/v1/namespaces/default/services/foo",
      },
      {
        type: getType(actions.kube.openWatchResource),
        payload: {
          ref,
          handler: expect.any(Function),
        },
      },
    ];

    store.dispatch(actions.kube.getAndWatchResource(ref));
    expect(store.getActions()).toEqual(expectedActions);
    expect(getResourceMock).toHaveBeenCalledWith("v1", "services", "default", "foo");
  });
});
