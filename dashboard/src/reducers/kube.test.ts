import { getType } from "typesafe-actions";
import actions from "../actions";
import ResourceRef from "../shared/ResourceRef";
import { IKubeState, IResource } from "../shared/types";
import kubeReducer from "./kube";

describe("authReducer", () => {
  let initialState: IKubeState;

  const actionTypes = {
    requestResource: getType(actions.kube.requestResource),
    receiveResource: getType(actions.kube.receiveResource),
    errorKube: getType(actions.kube.receiveResourceError),
    openWatchResource: getType(actions.kube.openWatchResource),
    closeWatchResource: getType(actions.kube.closeWatchResource),
  };

  const ref = new ResourceRef({
    apiVersion: "v1",
    kind: "Service",
    metadata: {
      name: "foo",
      namespace: "default",
    },
  } as IResource);

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

    describe("openWatchResource", () => {
      it("adds a new socket to the state for the requested resource", () => {
        const newState = kubeReducer(undefined, {
          type: actionTypes.openWatchResource,
          payload: {
            ref,
            handler: jest.fn(),
          },
        });
        const socket = newState.sockets[ref.watchResourceURL()];
        expect(socket).toBeDefined();
      });

      it("does not open a new socket if one exists in the state", () => {
        const existingSocket = ref.watchResource();
        const state = {
          ...initialState,
          sockets: {
            [ref.watchResourceURL()]: existingSocket,
          },
        };
        const newState = kubeReducer(state, {
          type: actionTypes.openWatchResource,
          payload: {
            ref,
            handler: jest.fn(),
          },
        });
        expect(newState).toBe(state);
        expect(newState.sockets[ref.watchResourceURL()]).toBe(existingSocket);
      });

      it("adds the requested handler on the created socket", () => {
        const mock = jest.fn();
        const newState = kubeReducer(undefined, {
          type: actionTypes.openWatchResource,
          payload: {
            ref,
            handler: mock,
          },
        });
        const socket = newState.sockets[ref.watchResourceURL()];
        // listeners is a defined property on the mock-socket:
        // https://github.com/thoov/mock-socket/blob/bed8c9237fa4b9c348a4cf5a22b59569c4cd10f2/index.d.ts#L7
        const listener = (socket as any).listeners.message[0];
        expect(listener).toBeDefined();
        listener();
        expect(mock).toHaveBeenCalled();
      });
    });

    describe("closeWatchResource", () => {
      it("closes the WebSocket for the requested resource and removes it from the state", () => {
        const socket = ref.watchResource();
        const spy = jest.spyOn(socket, "close");
        const state = {
          ...initialState,
          sockets: {
            [ref.watchResourceURL()]: socket,
          },
        };
        const newState = kubeReducer(state, {
          type: actionTypes.closeWatchResource,
          payload: ref,
        });
        expect(spy).toHaveBeenCalled();
        expect(newState.sockets).toEqual({});
      });
    });

    it("does nothing if the socket doesn't exist", () => {
      const state = { ...initialState, sockets: { dontdeleteme: {} as WebSocket } };
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
