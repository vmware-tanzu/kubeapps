import ResourceRef from "shared/ResourceRef";
import { IKubeState, IResource } from "shared/types";
import { getType } from "typesafe-actions";
import actions from "../actions";
import kubeReducer, { initialKinds } from "./kube";

const clusterName = "cluster-name";

describe("kubeReducer", () => {
  let initialState: IKubeState;

  const actionTypes = {
    requestResource: getType(actions.kube.requestResource),
    receiveResource: getType(actions.kube.receiveResource),
    errorKube: getType(actions.kube.receiveResourceError),
    openWatchResource: getType(actions.kube.openWatchResource),
    closeWatchResource: getType(actions.kube.closeWatchResource),
    receiveResourceFromList: getType(actions.kube.receiveResourceFromList),
    receiveResourceKinds: getType(actions.kube.receiveResourceKinds),
    requestResourceKinds: getType(actions.kube.requestResourceKinds),
    receiveKindsError: getType(actions.kube.receiveKindsError),
    addTimer: getType(actions.kube.addTimer),
    removeTimer: getType(actions.kube.removeTimer),
  };

  const ref = new ResourceRef(
    {
      apiVersion: "v1",
      kind: "Service",
      metadata: {
        name: "foo",
        namespace: "default",
      },
    } as IResource,
    clusterName,
    "services",
    true,
  );

  beforeEach(() => {
    initialState = {
      items: {},
      sockets: {},
      kinds: initialKinds,
      timers: {},
    };
  });

  describe("reducer actions", () => {
    it("request an item", () => {
      const payload = "foo";
      const type = actionTypes.requestResource as any;
      expect(kubeReducer(undefined, { type, payload })).toEqual({
        ...initialState,
        items: { foo: { isFetching: true } },
      });
    });
    it("receives an item", () => {
      const payload = { key: "foo", resource: { metadata: { name: "foo" } } as IResource };
      const type = actionTypes.receiveResource as any;
      expect(kubeReducer(undefined, { type, payload })).toEqual({
        ...initialState,
        items: { foo: { isFetching: false, item: { metadata: { name: "foo" } } } },
      });
    });
    it("receives an error", () => {
      const error = new Error("bar");
      const payload = { key: "foo", error };
      const type = actionTypes.errorKube as any;
      expect(kubeReducer(undefined, { type, payload })).toEqual({
        ...initialState,
        items: { foo: { isFetching: false, error } },
      });
    });

    it("receives an item from a list", () => {
      const resource = {
        metadata: { name: "foo", selfLink: "/foo" },
        status: { ready: false },
      } as IResource;
      const payload = { key: "foo", resource: { ...resource, status: { ready: true } } };
      const type = actionTypes.receiveResourceFromList as any;
      const stateWithItems = {
        ...initialState,
        items: { foo: { isFetching: true, item: { items: [resource] } } },
      } as any;
      expect(kubeReducer(stateWithItems, { type, payload })).toEqual({
        ...initialState,
        items: { foo: { isFetching: false, item: { items: [payload.resource] } } },
      });
    });

    it("receives an item when the list item is undefined", () => {
      const resource = {
        metadata: { name: "foo", selfLink: "/foo" },
        status: { ready: false },
      } as IResource;
      const payload = { key: "foo", resource: { ...resource, status: { ready: true } } };
      const type = actionTypes.receiveResourceFromList as any;
      const stateWithItems = {
        ...initialState,
        items: { foo: { isFetching: true, item: undefined } },
      } as any;
      expect(kubeReducer(stateWithItems, { type, payload })).toEqual({
        ...initialState,
        items: { foo: { isFetching: false, item: { items: [payload.resource] } } },
      });
    });

    describe("openWatchResource", () => {
      it("adds a new socket to the state for the requested resource", () => {
        const newState = kubeReducer(undefined, {
          type: actionTypes.openWatchResource,
          payload: {
            ref,
            handler: jest.fn(),
            onError: jest.fn(),
          },
        });
        const socket = newState.sockets[ref.watchResourceURL()];
        expect(socket).toBeDefined();
      });

      it("does not open a new socket if one exists in the state", () => {
        const existingSocket = ref.watchResource();
        const socket = { socket: existingSocket, onError: jest.fn() };
        const state = {
          ...initialState,
          sockets: {
            [ref.watchResourceURL()]: socket,
          },
        };
        const newState = kubeReducer(state, {
          type: actionTypes.openWatchResource,
          payload: {
            ref,
            handler: jest.fn(),
            onError: jest.fn(),
          },
        });
        expect(newState).toBe(state);
        expect(newState.sockets[ref.watchResourceURL()]).toBe(socket);
      });

      it("adds the requested handler on the created socket", () => {
        const mock = jest.fn();
        const newState = kubeReducer(undefined, {
          type: actionTypes.openWatchResource,
          payload: {
            ref,
            handler: mock,
            onError: jest.fn(),
          },
        });
        const socket = newState.sockets[ref.watchResourceURL()].socket;
        // listeners is a defined property on the mock-socket:
        // https://github.com/thoov/mock-socket/blob/bed8c9237fa4b9c348a4cf5a22b59569c4cd10f2/index.d.ts#L7
        const listener = (socket as any).listeners.message[0];
        expect(listener).toBeDefined();
        listener();
        expect(mock).toHaveBeenCalled();
      });

      it("triggers the onError function if the socket emits an error", () => {
        const mock = jest.fn();
        const newState = kubeReducer(undefined, {
          type: actionTypes.openWatchResource,
          payload: {
            ref,
            handler: jest.fn(),
            onError: mock,
          },
        });
        const socket = newState.sockets[ref.watchResourceURL()].socket;
        // listeners is a defined property on the mock-socket:
        // https://github.com/thoov/mock-socket/blob/bed8c9237fa4b9c348a4cf5a22b59569c4cd10f2/index.d.ts#L7
        const listener = (socket as any).listeners.error[0];
        expect(listener).toBeDefined();
        listener();
        expect(mock).toHaveBeenCalled();
      });
    });

    describe("closeWatchResource", () => {
      it("closes the WebSocket and the timer for the requested resource and removes it from the state", () => {
        const socket = ref.watchResource();
        const spy = jest.spyOn(socket, "close");
        socket.removeEventListener = jest.fn();
        const state = {
          ...initialState,
          sockets: {
            [ref.watchResourceURL()]: { socket, onError: jest.fn() },
          },
          timers: {
            [ref.getResourceURL()]: {} as NodeJS.Timer,
          },
        };
        const newState = kubeReducer(state, {
          type: actionTypes.closeWatchResource,
          payload: ref,
        });
        expect(spy).toHaveBeenCalled();
        expect(newState.sockets).toEqual({});
        expect(newState.timers).toEqual({ [ref.getResourceURL()]: undefined });
      });

      it("does nothing if the socket doesn't exist", () => {
        const state = {
          ...initialState,
          sockets: { dontdeleteme: { socket: {} as WebSocket, onError: jest.fn() } },
        };
        const newState = kubeReducer(state, {
          type: actionTypes.closeWatchResource,
          payload: ref,
        });
        expect(newState).toEqual(state);
        // check that dontdeleteme is not modified
        expect(newState.sockets.dontdeleteme).toBe(state.sockets.dontdeleteme);
      });
    });

    describe("addTimer", () => {
      it("should add a timer", () => {
        const newState = kubeReducer(initialState, {
          type: actionTypes.addTimer,
          payload: { id: "foo", timer: jest.fn() },
        });
        expect(newState).toEqual({
          ...initialState,
          timers: { foo: expect.any(Number) },
        });
      });

      it("should not add a timer if there is already one", () => {
        jest.useFakeTimers();
        const f1 = jest.fn();
        const f2 = jest.fn();
        const timer = setTimeout(f1, 1);
        const newState = kubeReducer(
          {
            ...initialState,
            timers: { foo: timer },
          },
          {
            type: actionTypes.addTimer,
            payload: { id: "foo", timer: f2 },
          },
        );
        expect(newState).toEqual({
          ...initialState,
          timers: { foo: timer },
        });
        jest.runAllTimers();
        expect(f1).toHaveBeenCalled();
        expect(f2).not.toHaveBeenCalled();
      });
    });

    describe("removeTimer", () => {
      it("remove a timer", () => {
        jest.useFakeTimers();
        const f1 = jest.fn();
        const timer = setTimeout(f1, 1);
        const newState = kubeReducer(
          {
            ...initialState,
            timers: { foo: timer },
          },
          {
            type: actionTypes.removeTimer,
            payload: "foo",
          },
        );
        expect(newState).toEqual({
          ...initialState,
          timers: { foo: undefined },
        });
        jest.runAllTimers();
        expect(f1).not.toHaveBeenCalled();
      });
    });

    describe("receiveResourceKinds", () => {
      it("contains default kinds", () => {
        const newState = kubeReducer(undefined, {
          type: actionTypes.requestResourceKinds,
        });
        expect(newState.kinds).toMatchObject({
          Deployment: { apiVersion: "apps/v1", plural: "deployments", namespaced: true },
        });
      });

      it("retrieve new kinds", () => {
        const kinds = {
          Deployment: {
            apiVersion: "apps/v1",
            plural: "deployments",
            namespaced: true,
          },
        };
        const newState = kubeReducer(undefined, {
          type: actionTypes.receiveResourceKinds,
          payload: kinds,
        });
        expect(newState).toEqual({
          ...initialState,
          kinds,
        });
      });

      it("sets an error", () => {
        const newState = kubeReducer(undefined, {
          type: actionTypes.receiveKindsError,
          payload: new Error("nope!"),
        });
        expect(newState).toEqual({
          ...initialState,
          kindsError: new Error("nope!"),
        });
      });
    });
  });
});
