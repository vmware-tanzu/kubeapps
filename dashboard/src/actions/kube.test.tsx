import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";
import { Kube } from "shared/Kube";
import { IKubeState, IResource } from "shared/types";
import { getType } from "typesafe-actions";
import actions from ".";
import {
  InstalledPackageReference,
  ResourceRef as APIResourceRef,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { GetResourcesResponse } from "gen/kubeappsapis/plugins/resources/v1alpha1/resources";

const mockStore = configureMockStore([thunk]);

let store: any;

beforeEach(() => {
  store = mockStore({
    kube: {
      items: {},
      subscriptions: {},
      kinds: {},
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

describe("getResourceKinds", () => {
  beforeEach(() => {
    Kube.getAPIGroups = jest.fn().mockResolvedValue([]);
    Kube.getResourceKinds = jest.fn().mockResolvedValue({});
  });
  afterEach(() => {
    jest.restoreAllMocks();
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

describe("getResources", () => {
  it("dispatches a requestResources action", () => {
    const refs = [
      {
        apiVersion: "v1",
        kind: "Service",
        name: "foo",
        namespace: "default",
      },
    ] as APIResourceRef[];

    const pkg = {
      identifier: "test-pkg",
    } as InstalledPackageReference;

    const watch = true;

    const expectedActions = [
      {
        type: getType(actions.kube.requestResources),
        payload: {
          pkg,
          refs,
          watch,
          handler: expect.any(Function),
          onError: expect.any(Function),
          onComplete: expect.any(Function),
        },
      },
    ];

    store.dispatch(actions.kube.getResources(pkg, refs, watch));
    expect(store.getActions()).toEqual(expectedActions);
  });
});

describe("processGetResourcesResponse", () => {
  it("dispatches a receiveResource action with the expected key", () => {
    const expectedResource = {
      apiVersion: "v1",
      kind: "Service",
      metadata: {
        name: "foo",
        namespace: "default",
      },
    } as IResource;

    const getResourcesResponse = {
      resourceRef: {
        apiVersion: "v1",
        kind: "Service",
        name: "foo",
        namespace: "default",
      } as APIResourceRef,
      manifest: {
        value: new TextEncoder().encode(JSON.stringify(expectedResource)),
        typeUrl: "",
      },
    } as GetResourcesResponse;

    const expectedKey = "v1/Service/default/foo";

    const expectedActions = [
      {
        type: getType(actions.kube.receiveResource),
        payload: {
          key: expectedKey,
          resource: expectedResource,
        },
      },
    ];

    actions.kube.processGetResourcesResponse(getResourcesResponse, store.dispatch);

    expect(store.getActions()).toEqual(expectedActions);
  });

  it("dispatches an error action if the resourceRef is missing", () => {
    const expectedResource = {
      apiVersion: "v1",
      kind: "Service",
      metadata: {
        name: "foo",
        namespace: "default",
      },
    } as IResource;

    const getResourcesResponse = {
      manifest: {
        value: new TextEncoder().encode(JSON.stringify(expectedResource)),
        typeUrl: "",
      },
    } as GetResourcesResponse;

    const expectedActions = [
      {
        type: getType(actions.kube.receiveResourcesError),
        payload: new Error("received resource without a resource reference"),
      },
    ];

    actions.kube.processGetResourcesResponse(getResourcesResponse, store.dispatch);

    expect(store.getActions()).toEqual(expectedActions);
  });
});
