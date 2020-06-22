import { LOCATION_CHANGE, RouterActionType } from "connected-react-router";
import context from "jest-plugin-context";
import { getType } from "typesafe-actions";

import actions from "../actions";
import { IResource } from "../shared/types";
import namespaceReducer from "./namespace";

describe("namespaceReducer", () => {
  const initialState = {
    cluster: "default",
    current: "initial-current",
    namespaces: ["default", "initial-current"],
  };
  context("when LOCATION CHANGE", () => {
    const location = {
      hash: "",
      search: "",
      state: "",
    };

    describe("changes the current stored namespace if it is in the URL", () => {
      const testCases = [
        { path: "/c/default/ns/cyberdyne/apps", current: "cyberdyne" },
        { path: "/cyberdyne/apps", current: "initial-current" },
        { path: "/c/barcluster/ns/T-600/charts", current: "T-600" },
      ];
      testCases.forEach(tc => {
        it(tc.path, () =>
          expect(
            namespaceReducer(initialState, {
              type: LOCATION_CHANGE,
              payload: {
                location: { ...location, pathname: tc.path },
                action: "PUSH" as RouterActionType,
              },
            }),
          ).toEqual({ ...initialState, current: tc.current }),
        );
      });
    });
  });

  context("when ERROR_NAMESPACE", () => {
    const err = new Error("Bang!");

    it("when listing leaves namespaces intact and but ignores the error", () => {
      expect(
        namespaceReducer(initialState, {
          type: getType(actions.namespace.errorNamespaces),
          payload: { err, op: "list" },
        }),
      ).toEqual({ ...initialState, error: undefined });
    });

    it("leaves namespaces intact and sets the error", () => {
      expect(
        namespaceReducer(initialState, {
          type: getType(actions.namespace.errorNamespaces),
          payload: { err, op: "create" },
        }),
      ).toEqual({ ...initialState, error: { action: "create", error: err } });
    });
  });

  context("when CLEAR_NAMESPACES", () => {
    it("returns to the initial state", () => {
      const dirtyState = {
        cluster: "default",
        current: "some-other-namespace",
        namespaces: ["namespace-one", "namespace-two"],
      };
      expect(
        namespaceReducer(dirtyState, {
          type: getType(actions.namespace.clearNamespaces),
        }),
      ).toEqual({ cluster: "default", current: "_all", namespaces: [] });
    });
  });

  context("when SET_AUTHENTICATED", () => {
    it("sets the current namespace to the users default", () => {
      expect(
        namespaceReducer(
          { cluster: "default", current: "default", namespaces: [] },
          {
            type: getType(actions.auth.setAuthenticated),
            payload: { authenticated: true, oidc: false, defaultNamespace: "foo-bar" },
          },
        ),
      ).toEqual({ cluster: "default", current: "foo-bar", namespaces: [] });
    });
  });

  context("when SET_NAMESPACE", () => {
    it("sets the current namespace and clears error", () => {
      expect(
        namespaceReducer(
          {
            cluster: "default",
            current: "error",
            namespaces: [],
            error: { action: "create", error: new Error("boom") },
          },
          {
            type: getType(actions.namespace.setNamespace),
            payload: "default",
          },
        ),
      ).toEqual({ cluster: "default", current: "default", namespaces: [], error: undefined });
    });
  });

  context("when RECEIVE_NAMESPACE", () => {
    it("adds the namespace to the list and clears error", () => {
      expect(
        namespaceReducer(
          {
            cluster: "default",
            current: "",
            namespaces: ["default"],
            error: { action: "create", error: new Error("boom") },
          },
          {
            type: getType(actions.namespace.receiveNamespace),
            payload: { metadata: { name: "bar" } } as IResource,
          },
        ),
      ).toEqual({ cluster: "default", current: "", namespaces: ["bar", "default"], error: undefined });
    });
  });
});
