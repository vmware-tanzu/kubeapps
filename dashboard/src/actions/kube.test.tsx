// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import {
  InstalledPackageReference,
  ResourceRef as APIResourceRef,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { GetResourcesResponse } from "gen/kubeappsapis/plugins/resources/v1alpha1/resources";
import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";
import { Kube } from "shared/Kube";
import { IKubeState, IResource, IStoreState } from "shared/types";
import { getType } from "typesafe-actions";
import actions from ".";

const mockStore = configureMockStore([thunk]);

const makeStore = (operatorsEnabled: boolean) => {
  const state: IKubeState = {
    items: {},
    subscriptions: {},
    kinds: {},
  };
  const config = operatorsEnabled ? { featureFlags: { operators: true } } : {};
  return mockStore({ kube: state, config: config } as Partial<IStoreState>);
};

let store: any;

beforeEach(() => {
  store = makeStore(false);
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
    const groups = [{ name: "foo" }, { name: "operators.coreos.com" }];
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
    const groups = [{ name: "foo" }];
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
  it("dispatches a requestResources action with an onComplete that closes request when watching", () => {
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
      manifest: JSON.stringify(expectedResource),
    } as GetResourcesResponse;

    const expectedHandlerActions = [
      expectedActions[0],
      {
        type: getType(actions.kube.receiveResource),
        payload: {
          key: "v1/Service/default/foo",
          resource: expectedResource,
        },
      },
    ];
    const handler = store.getActions()[0].payload.handler;
    handler(getResourcesResponse);
    expect(store.getActions()).toEqual(expectedHandlerActions);

    const expectedCompletionActions = [
      ...expectedHandlerActions,
      {
        type: getType(actions.kube.closeRequestResources),
        payload: pkg,
      },
    ];
    const onComplete = store.getActions()[0].payload.onComplete;
    onComplete();
    expect(store.getActions()).toEqual(expectedCompletionActions);
  });

  it("dispatches a requestResources action with an onComplete that does nothing when not watching", () => {
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

    const watch = false;

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

    const onComplete = store.getActions()[0].payload.onComplete;
    onComplete();
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
      manifest: JSON.stringify(expectedResource),
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
      manifest: JSON.stringify(expectedResource),
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
