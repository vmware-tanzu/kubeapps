// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";
import { Kube } from "shared/Kube";
import Namespace from "shared/Namespace";
import { getType } from "typesafe-actions";
import {
  canCreate,
  createNamespace,
  errorNamespaces,
  fetchNamespaces,
  checkNamespaceExists,
  postNamespace,
  receiveNamespaceExists,
  receiveNamespaces,
  requestNamespaceExists,
  setAllowCreate,
  setNamespace,
  setNamespaceState,
} from "./namespace";

const mockStore = configureMockStore([thunk]);

let store: any;

beforeEach(() => {
  store = mockStore();
});
afterEach(() => {
  jest.resetAllMocks();
});

// Regular action creators
// Regular action creators
interface ITestCase {
  name: string;
  action: (...args: any[]) => any;
  args?: any;
  payload?: any;
}

const actionTestCases: ITestCase[] = [
  {
    name: "setNamespace",
    action: setNamespaceState,
    args: ["default", "jack"],
    payload: { cluster: "default", namespace: "jack" },
  },
  {
    name: "receiveNamespces",
    action: receiveNamespaces,
    args: ["default", ["jack", "danny"]],
    payload: { cluster: "default", namespaces: ["jack", "danny"] },
  },
];

actionTestCases.forEach(tc => {
  describe(tc.name, () => {
    it("has expected structure", () => {
      expect(tc.action.call(null, ...tc.args)).toEqual({
        type: getType(tc.action),
        payload: tc.payload,
      });
    });
  });
});

// Async action creators
describe("fetchNamespaces", () => {
  it("dispatches the list of namespace names if no error", async () => {
    Namespace.list = jest.fn().mockImplementationOnce(() => {
      return ["overlook-hotel", "room-217"];
    });
    const expectedActions = [
      {
        type: getType(receiveNamespaces),
        payload: { cluster: "default-c", namespaces: ["overlook-hotel", "room-217"] },
      },
    ];

    await store.dispatch(fetchNamespaces("default-c"));
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("dispatches errorNamespace if the request returns no 'namespaces'", async () => {
    Namespace.list = jest.fn().mockImplementationOnce(() => {
      return [];
    });
    const err = new Error("The current account does not have access to any namespaces");
    const expectedActions = [
      {
        type: getType(errorNamespaces),
        payload: { cluster: "default-c", err, op: "list" },
      },
    ];

    await store.dispatch(fetchNamespaces("default-c"));
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("dispatches errorNamespace if error listing namespaces", async () => {
    const err = new Error("Bang!");
    Namespace.list = jest.fn().mockImplementationOnce(() => Promise.reject(err));
    const expectedActions = [
      {
        type: getType(errorNamespaces),
        payload: { cluster: "default-c", err, op: "list" },
      },
    ];

    await store.dispatch(fetchNamespaces("default-c"));

    expect(store.getActions()).toEqual(expectedActions);
  });
});

describe("createNamespace", () => {
  it("dispatches the new namespace and re-fetch namespaces", async () => {
    Namespace.create = jest.fn();
    Namespace.list = jest.fn().mockImplementationOnce(() => {
      return ["overlook-hotel", "room-217"];
    });
    const expectedActions = [
      {
        type: getType(postNamespace),
        payload: { cluster: "default-c", namespace: "overlook-hotel" },
      },
      {
        type: getType(receiveNamespaces),
        payload: { cluster: "default-c", namespaces: ["overlook-hotel", "room-217"] },
      },
    ];

    const res = await store.dispatch(createNamespace("default-c", "overlook-hotel", {}));
    expect(res).toBe(true);
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("dispatches errorNamespace if error getting a namespace", async () => {
    const err = new Error("Bang!");
    Namespace.create = jest.fn().mockImplementationOnce(() => Promise.reject(err));
    const expectedActions = [
      {
        type: getType(errorNamespaces),
        payload: { cluster: "default-c", err, op: "create" },
      },
    ];

    const res = await store.dispatch(createNamespace("default-c", "foo", {}));
    expect(res).toBe(false);
    expect(store.getActions()).toEqual(expectedActions);
  });
});

describe("checkNamespaceExists", () => {
  it("returns true if the namespace exists", async () => {
    Namespace.exists = jest.fn().mockReturnValue(true);
    const expectedActions = [
      {
        type: getType(requestNamespaceExists),
        payload: { cluster: "default-c", namespace: "default-ns" },
      },
      {
        type: getType(receiveNamespaceExists),
        payload: { cluster: "default-c", namespace: "default-ns" },
      },
    ];

    const exists = await store.dispatch(checkNamespaceExists("default-c", "default-ns"));

    expect(exists).toBe(true);
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("returns false if the namespace does not exist", async () => {
    Namespace.exists = jest.fn().mockReturnValue(false);
    const expectedActions = [
      {
        type: getType(requestNamespaceExists),
        payload: { cluster: "default-c", namespace: "default-ns" },
      },
    ];

    const exists = await store.dispatch(checkNamespaceExists("default-c", "default-ns"));

    expect(exists).toBe(false);
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("dispatches errorNamespace if error creating a namespace", async () => {
    const err = new Error("Bang!");
    Namespace.exists = jest.fn().mockImplementationOnce(() => Promise.reject(err));
    const expectedActions = [
      {
        type: getType(requestNamespaceExists),
        payload: { cluster: "default-c", namespace: "default-ns" },
      },
      {
        type: getType(errorNamespaces),
        payload: { cluster: "default-c", err, op: "get" },
      },
    ];
    const r = await store.dispatch(checkNamespaceExists("default-c", "default-ns"));
    expect(r).toBe(false);
    expect(store.getActions()).toEqual(expectedActions);
  });
});

describe("setNamespace", () => {
  it("dispatches namespace set", async () => {
    const expectedActions = [
      {
        type: getType(setNamespaceState),
        payload: { cluster: "default-c", namespace: "default-ns" },
      },
    ];
    await store.dispatch(setNamespace("default-c", "default-ns"));
    expect(store.getActions()).toEqual(expectedActions);
    expect(localStorage.setItem).toHaveBeenCalledWith(
      "kubeapps_namespace",
      '{"default-c":"default-ns"}',
    );
  });
});

describe("canCreate", () => {
  it("checks if it can create namespaces", async () => {
    Kube.canI = jest.fn().mockReturnValue({
      then: jest.fn(f => f(true)),
      catch: jest.fn(f => f(false)),
    });
    const expectedActions = [
      {
        type: getType(setAllowCreate),
        payload: { cluster: "default-c", allowed: true },
      },
    ];
    await store.dispatch(canCreate("default-c"));
    expect(store.getActions()).toEqual(expectedActions);
    expect(Kube.canI).toHaveBeenCalledWith("default-c", "", "namespaces", "create", "");
  });

  it("dispatches an error", async () => {
    const err = new Error("boom");
    Kube.canI = jest.fn(() => {
      throw err;
    });
    const expectedActions = [
      {
        type: getType(errorNamespaces),
        payload: { cluster: "default-c", err, op: "get" },
      },
    ];
    await store.dispatch(canCreate("default-c"));
    expect(store.getActions()).toEqual(expectedActions);
  });
});
