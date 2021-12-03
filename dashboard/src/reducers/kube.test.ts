import { IKubeState, IResource } from "shared/types";
import { getType } from "typesafe-actions";
import actions from "../actions";
import kubeReducer, { initialKinds } from "./kube";
import {
  Context,
  InstalledPackageReference,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { Kube } from "shared/Kube";

describe("kubeReducer", () => {
  let initialState: IKubeState;

  const actionTypes = {
    receiveResource: getType(actions.kube.receiveResource),
    errorKube: getType(actions.kube.receiveResourceError),
    receiveResourceKinds: getType(actions.kube.receiveResourceKinds),
    requestResourceKinds: getType(actions.kube.requestResourceKinds),
    receiveKindsError: getType(actions.kube.receiveKindsError),
    addTimer: getType(actions.kube.addTimer),
    removeTimer: getType(actions.kube.removeTimer),
  };

  beforeEach(() => {
    initialState = {
      items: {},
      sockets: {},
      subscriptions: {},
      kinds: initialKinds,
      timers: {},
    };
  });

  describe("reducer actions", () => {
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

    describe("requestResources", () => {
      // Ensure our shared/Kube helper is not calling out on the network.
      jest.mock("shared/Kube");

      const pkg = {
        context: {
          cluster: "default",
          namespace: "test",
        } as Context,
        identifier: "test-pkg",
      } as InstalledPackageReference;
      const key = `${pkg.context?.cluster}/${pkg.context?.namespace}/${pkg.identifier}`;

      const defaultPayload = {
        pkg,
        refs: [],
        watch: false,
        handler: jest.fn(),
        onError: jest.fn(),
        onComplete: jest.fn(),
      };

      it("adds a new subscription to the state for the requested package", () => {
        const newState = kubeReducer(undefined, {
          type: getType(actions.kube.requestResources),
          payload: defaultPayload,
        });

        expect(newState.subscriptions[key]).toBeDefined();
      });

      it("does not create a new subscription if one exists in the state", () => {
        const subscription = Kube.getResources(pkg, [], true).subscribe({});
        const state = {
          ...initialState,
          subscriptions: {
            [key]: subscription,
          },
        };
        const newState = kubeReducer(state, {
          type: getType(actions.kube.requestResources),
          payload: defaultPayload,
        });
        expect(newState).toBe(state);
        expect(newState.subscriptions[key]).toBe(subscription);
      });
    });
  });
});
