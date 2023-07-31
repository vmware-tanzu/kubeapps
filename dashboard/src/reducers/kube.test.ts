// Copyright 2018-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { IKubeState, IResource } from "shared/types";
import { getType } from "typesafe-actions";
import actions from "../actions";
import kubeReducer, { initialKinds } from "./kube";

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
  });
});
