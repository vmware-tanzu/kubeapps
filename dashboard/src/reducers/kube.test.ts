// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

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
  };

  beforeEach(() => {
    initialState = {
      items: {},
      subscriptions: {},
      kinds: initialKinds,
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

      const subscribe = jest.fn(() => true);

      beforeEach(() => {
        const observable = {
          subscribe,
        };
        Kube.getResources = jest.fn(() => observable as any);
      });

      afterEach(() => {
        jest.resetAllMocks();
      });

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

      it("adds a new subscription to the state when watching the requested package", () => {
        const newState = kubeReducer(undefined, {
          type: getType(actions.kube.requestResources),
          payload: {
            ...defaultPayload,
            watch: true,
          },
        });

        expect(Kube.getResources).toBeCalledWith(pkg, [], true);
        expect(subscribe).toHaveBeenCalled();
        expect(newState.subscriptions[key]).toBeDefined();
      });

      it("does not add a new subscription to the state when getting the requested package", () => {
        const newState = kubeReducer(undefined, {
          type: getType(actions.kube.requestResources),
          payload: {
            ...defaultPayload,
            watch: false,
          },
        });

        expect(Kube.getResources).toBeCalledWith(pkg, [], false);
        expect(subscribe).toHaveBeenCalled();
        expect(newState.subscriptions[key]).toBeUndefined();
      });
    });
  });
});
