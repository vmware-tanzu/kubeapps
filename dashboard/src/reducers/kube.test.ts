import { getType } from "typesafe-actions";
import actions from "../actions";
import ResourceRef from "../shared/ResourceRef";
import { IKubeState, IResource } from "../shared/types";
import kubeReducer from "./kube";

const clusterName = "cluster-name";

describe("authReducer", () => {
  let initialState: IKubeState;

  const actionTypes = {
    requestResource: getType(actions.kube.requestResource),
    receiveResource: getType(actions.kube.receiveResource),
    errorKube: getType(actions.kube.receiveResourceError),
    openWatchResource: getType(actions.kube.openWatchResource),
    closeWatchResource: getType(actions.kube.closeWatchResource),
    receiveResourceFromList: getType(actions.kube.receiveResourceFromList),
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
  );

  beforeEach(() => {
    initialState = {
      items: {},
      sockets: {},
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
            onError: { onErrorHandler: jest.fn(), closeTimer: jest.fn() },
          },
        });
        const socket = newState.sockets[ref.watchResourceURL()];
        expect(socket).toBeDefined();
      });

      it("does not open a new socket if one exists in the state", () => {
        const existingSocket = ref.watchResource();
        const socket = { socket: existingSocket, closeTimer: jest.fn() };
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
            onError: { onErrorHandler: jest.fn(), closeTimer: jest.fn() },
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
            onError: { onErrorHandler: jest.fn(), closeTimer: jest.fn() },
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
            onError: { onErrorHandler: mock, closeTimer: jest.fn() },
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
        const timerMock = jest.fn();
        const spy = jest.spyOn(socket, "close");
        const state = {
          ...initialState,
          sockets: {
            [ref.watchResourceURL()]: { socket, closeTimer: timerMock },
          },
        };
        const newState = kubeReducer(state, {
          type: actionTypes.closeWatchResource,
          payload: ref,
        });
        expect(spy).toHaveBeenCalled();
        expect(newState.sockets).toEqual({});
        expect(timerMock).toHaveBeenCalled();
      });
    });

    it("does nothing if the socket doesn't exist", () => {
      const state = {
        ...initialState,
        sockets: { dontdeleteme: { socket: {} as WebSocket, closeTimer: jest.fn() } },
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
});
