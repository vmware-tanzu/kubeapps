import { LOCATION_CHANGE, RouterActionType } from "connected-react-router";
import context from "jest-plugin-context";
import { getType } from "typesafe-actions";

import actions from "../actions";
import { IResource } from "../shared/types";
import clusterReducer, { IClustersState } from "./cluster";

describe("namespaceReducer", () => {
  const initialState: IClustersState = {
    currentCluster: "default",
    clusters: {
      default: {
        currentNamespace: "initial-current",
        namespaces: ["default", "initial-current"],
      },
    },
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
            clusterReducer(initialState, {
              type: LOCATION_CHANGE,
              payload: {
                location: { ...location, pathname: tc.path },
                action: "PUSH" as RouterActionType,
              },
            }),
          ).toEqual({
            ...initialState,
            clusters: {
              default: {
                ...initialState.clusters.default,
                currentNamespace: tc.current,
              },
            },
          } as IClustersState)
        );
      });
    });
  });

  context("when ERROR_NAMESPACE", () => {
    const err = new Error("Bang!");

    it("when listing leaves namespaces intact and but ignores the error", () => {
      expect(
        clusterReducer(initialState, {
          type: getType(actions.namespace.errorNamespaces),
          payload: { err, op: "list" },
        }),
      ).toEqual({
        ...initialState,
        clusters: {
          default: {
            ...initialState.clusters.default,
            error: undefined,
          },
        },
      } as IClustersState);
    });

    it("leaves namespaces intact and sets the error", () => {
      expect(
        clusterReducer(initialState, {
          type: getType(actions.namespace.errorNamespaces),
          payload: { err, op: "create" },
        }),
      ).toEqual({
        ...initialState,
        clusters: {
          default: { ...initialState.clusters.default, error: { action: "create", error: err } }
        },
      } as IClustersState);
    });
  });

  context("when CLEAR_NAMESPACES", () => {
    it("returns to the initial state", () => {
      expect(
        clusterReducer(initialState, {
          type: getType(actions.namespace.clearNamespaces),
        }),
      ).toEqual({
        ...initialState,
        clusters: { default: { currentNamespace: "_all", namespaces: [] } }
      } as IClustersState);
    });
  });

  context("when SET_AUTHENTICATED", () => {
    it("sets the current namespace to the users default", () => {
      expect(
        clusterReducer(
          initialState,
          {
            type: getType(actions.auth.setAuthenticated),
            payload: { authenticated: true, oidc: false, defaultNamespace: "foo-bar" },
          },
        ),
      ).toEqual({
        ...initialState,
        clusters: {
          default: { ...initialState.clusters.default, currentNamespace: "foo-bar" },
        },
      } as IClustersState);
    });
  });

  context("when SET_NAMESPACE", () => {
    it("sets the current namespace and clears error", () => {
      expect(
        clusterReducer({
          ...initialState,
          clusters: {
            default: {
              ...initialState.clusters.default,
              currentNamespace: "other",
              error: { action: "create", error: new Error("Bang!") },
            },
          },
        }, {
          type: getType(actions.namespace.setNamespace),
          payload: "default",
        }),
      ).toEqual({
        ...initialState,
        clusters: {
          default: {
            ...initialState.clusters.default,
            currentNamespace: "default",
            error: undefined,
          },
        },
      } as IClustersState);
    });
  });

  context("when RECEIVE_NAMESPACE", () => {
    it("adds the namespace to the list and clears error", () => {
      expect(
        clusterReducer(
          {
            ...initialState,
            clusters: {
              default: {
                currentNamespace: "",
                namespaces: ["default"],
                error: { action: "create", error: new Error("boom") },
              }
            },
          } as IClustersState,
          {
            type: getType(actions.namespace.receiveNamespace),
            payload: { metadata: { name: "bar" } } as IResource,
          },
        ),
      ).toEqual({
        ...initialState,
        clusters: {
          default: {
            currentNamespace: "",
            namespaces: ["bar", "default"],
            error: undefined,
          },
        },
      } as IClustersState);
    });
  });
});
