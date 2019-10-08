import { LOCATION_CHANGE, RouterActionType } from "connected-react-router";
import context from "jest-plugin-context";
import { getType } from "typesafe-actions";

import actions from "../actions";
import namespaceReducer from "./namespace";

describe("namespaceReducer", () => {
  const initialState = {
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
        { path: "/ns/cyberdyne/apps", current: "cyberdyne" },
        { path: "/cyberdyne/apps", current: "initial-current" },
        { path: "/ns/T-600/charts", current: "T-600" },
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

    it("leaves namespaces intact and sets error", () => {
      expect(
        namespaceReducer(initialState, {
          type: getType(actions.namespace.errorNamespaces),
          payload: { err, op: "list" },
        }),
      ).toEqual({ ...initialState, errorMsg: err.message });
    });
  });

  context("when CLEAR_NAMESPACES", () => {
    it("returns to the initial state", () => {
      const dirtyState = {
        current: "some-other-namespace",
        namespaces: ["namespace-one", "namespace-two"],
      };
      expect(
        namespaceReducer(dirtyState, {
          type: getType(actions.namespace.clearNamespaces),
        }),
      ).toEqual({ current: "default", namespaces: [] });
    });
  });

  context("when SET_AUTHENTICATED", () => {
    it("sets the current namespace to the users default", () => {
      expect(
        namespaceReducer(
          { current: "default", namespaces: [] },
          {
            type: getType(actions.auth.setAuthenticated),
            payload: { authenticated: true, oidc: false, defaultNamespace: "foo-bar" },
          },
        ),
      ).toEqual({ current: "foo-bar", namespaces: [] });
    });
  });
});
