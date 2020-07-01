import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";
import { getType } from "typesafe-actions";
import Namespace from "../shared/Namespace";
import {
  createNamespace,
  errorNamespaces,
  fetchNamespaces,
  getNamespace,
  postNamespace,
  receiveNamespace,
  receiveNamespaces,
  requestNamespace,
  setNamespace,
} from "./namespace";

const mockStore = configureMockStore([thunk]);

let store: any;

beforeEach(() => {
  store = mockStore();
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
  { name: "setNamespace", action: setNamespace, args: "jack", payload: "jack" },
  {
    name: "receiveNamespces",
    action: receiveNamespaces,
    args: ["jack", "danny"],
    payload: ["jack", "danny"],
  },
];

actionTestCases.forEach(tc => {
  describe(tc.name, () => {
    it("has expected structure", () => {
      expect(tc.action.call(null, tc.args)).toEqual({
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
      return {
        namespaces: [{ metadata: { name: "overlook-hotel" } }, { metadata: { name: "room-217" } }],
      };
    });
    const expectedActions = [
      {
        type: getType(receiveNamespaces),
        payload: ["overlook-hotel", "room-217"],
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
        payload: { err, op: "list" },
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
      return {
        namespaces: [{ metadata: { name: "overlook-hotel" } }, { metadata: { name: "room-217" } }],
      };
    });
    const expectedActions = [
      {
        type: getType(postNamespace),
        payload: "overlook-hotel",
      },
      {
        type: getType(receiveNamespaces),
        payload: ["overlook-hotel", "room-217"],
      },
    ];

    const res = await store.dispatch(createNamespace("default-c", "overlook-hotel"));
    expect(res).toBe(true);
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("dispatches errorNamespace if error getting a namespace", async () => {
    const err = new Error("Bang!");
    Namespace.create = jest.fn().mockImplementationOnce(() => Promise.reject(err));
    const expectedActions = [
      {
        type: getType(errorNamespaces),
        payload: { err, op: "create" },
      },
    ];

    const res = await store.dispatch(createNamespace("default-c", "foo"));
    expect(res).toBe(false);
    expect(store.getActions()).toEqual(expectedActions);
  });
});

describe("getNamespace", () => {
  it("dispatches requested namespace", async () => {
    const ns = { metadata: { name: "default" } };
    Namespace.get = jest.fn().mockReturnValue(ns);
    const expectedActions = [
      {
        type: getType(requestNamespace),
        payload: "default-ns",
      },
      {
        type: getType(receiveNamespace),
        payload: ns,
      },
    ];
    const r = await store.dispatch(getNamespace("default-c", "default-ns"));
    expect(r).toBe(true);
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("dispatches errorNamespace if error creating a namespace", async () => {
    const err = new Error("Bang!");
    Namespace.get = jest.fn().mockImplementationOnce(() => Promise.reject(err));
    const expectedActions = [
      {
        type: getType(requestNamespace),
        payload: "default-ns",
      },
      {
        type: getType(errorNamespaces),
        payload: { err, op: "get" },
      },
    ];
    const r = await store.dispatch(getNamespace("default-c", "default-ns"));
    expect(r).toBe(false);
    expect(store.getActions()).toEqual(expectedActions);
  });
});
