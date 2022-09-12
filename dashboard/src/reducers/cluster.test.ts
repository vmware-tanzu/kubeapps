// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { LOCATION_CHANGE, RouterActionType, RouterLocation } from "connected-react-router";
import { Location } from "history";
import context from "jest-plugin-context";
import { Auth } from "shared/Auth";
import { IConfig } from "shared/Config";
import { getType } from "typesafe-actions";
import actions from "../actions";
import clusterReducer, { IClustersState, initialState } from "./cluster";

describe("clusterReducer", () => {
  const initialTestState: IClustersState = {
    currentCluster: "initial-cluster",
    clusters: {
      default: {
        currentNamespace: "initial-namespace",
        namespaces: ["default", "initial-namespace"],
        canCreateNS: true,
      },
      "initial-cluster": {
        currentNamespace: "initial-namespace",
        namespaces: ["default", "initial-namespace"],
        canCreateNS: true,
      },
    },
  };
  context("when LOCATION CHANGE", () => {
    const location: Location = {
      hash: "",
      pathname: "",
      search: "",
      state: "",
      key: "",
    };

    describe("changes the current stored namespace if it is in the URL", () => {
      const testCases = [
        {
          name: "updates both the cluster and namespace",
          path: "/c/default/ns/cyberdyne/apps",
          currentNamespace: "cyberdyne",
          currentCluster: "default",
        },
        {
          name: "does not change cluster or namespace when neither is in url",
          path: "/cyberdyne/apps",
          currentNamespace: "initial-namespace",
          currentCluster: "initial-cluster",
        },
        {
          name: "updates namespace for a non-multicluster URI",
          path: "/ns/default/operators",
          currentNamespace: "default",
          currentCluster: "initial-cluster",
        },
        {
          name: "updates cluster for a non-namespaced URI",
          path: "/c/default/config/brokers",
          currentNamespace: "initial-namespace",
          currentCluster: "default",
        },
      ];
      testCases.forEach(tc => {
        it(tc.name, () =>
          expect(
            clusterReducer(initialTestState, {
              type: LOCATION_CHANGE,
              payload: {
                location: { ...location, pathname: tc.path } as RouterLocation<any>,
                action: "PUSH" as RouterActionType,
                isFirstRendering: true,
              },
            }),
          ).toEqual({
            ...initialTestState,
            currentCluster: tc.currentCluster,
            clusters: {
              ...initialTestState.clusters,
              [tc.currentCluster]: {
                ...initialTestState.clusters[tc.currentCluster],
                currentNamespace: tc.currentNamespace,
              },
            },
          } as IClustersState),
        );
      });
    });
  });

  context("when ERROR_NAMESPACE", () => {
    const err = new Error("Bang!");
    it("leaves namespaces intact and sets the error", () => {
      expect(
        clusterReducer(initialTestState, {
          type: getType(actions.namespace.errorNamespaces),
          payload: { cluster: "initial-cluster", err, op: "create" },
        }),
      ).toEqual({
        ...initialTestState,
        clusters: {
          ...initialTestState.clusters,
          "initial-cluster": {
            ...initialTestState.clusters["initial-cluster"],
            error: { action: "create", error: err },
          },
        },
      } as IClustersState);
    });
  });

  context("when CLEAR_CLUSTERS", () => {
    it("returns to the initial state", () => {
      expect(
        clusterReducer(initialTestState, {
          type: getType(actions.namespace.clearClusters),
        }),
      ).toEqual({
        ...initialTestState,
        clusters: {
          ...initialState.clusters,
        },
      } as IClustersState);
    });
  });

  context("when SET_NAMESPACE", () => {
    it("sets the current namespace and clears error", () => {
      expect(
        clusterReducer(
          {
            ...initialTestState,
            clusters: {
              ...initialTestState.clusters,
              "initial-cluster": {
                ...initialTestState.clusters["initial-cluster"],
                currentNamespace: "other",
                error: { action: "create", error: new Error("Bang!") },
              },
            },
          },
          {
            type: getType(actions.namespace.setNamespaceState),
            payload: { cluster: "initial-cluster", namespace: "default" },
          },
        ),
      ).toEqual({
        ...initialTestState,
        clusters: {
          ...initialTestState.clusters,
          "initial-cluster": {
            ...initialTestState.clusters["initial-cluster"],
            currentNamespace: "default",
            error: undefined,
            canCreateNS: true,
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
            ...initialTestState,
            clusters: {
              default: {
                currentNamespace: "",
                namespaces: ["default"],
                canCreateNS: true,
              },
              other: {
                currentNamespace: "",
                namespaces: ["othernamespace"],
                error: { action: "create", error: new Error("boom") },
                canCreateNS: true,
              },
            },
          } as IClustersState,
          {
            type: getType(actions.namespace.receiveNamespaceExists),
            payload: {
              cluster: "other",
              namespace: "bar",
            },
          },
        ),
      ).toEqual({
        ...initialTestState,
        clusters: {
          default: {
            currentNamespace: "",
            namespaces: ["default"],
            canCreateNS: true,
          },
          other: {
            currentNamespace: "",
            namespaces: ["bar", "othernamespace"],
            error: undefined,
            canCreateNS: true,
          },
        },
      } as IClustersState);
    });
  });

  context("when RECEIVE_NAMESPACES", () => {
    afterEach(() => {
      jest.restoreAllMocks();
    });

    it("updates the namespace list and clears error", () => {
      expect(
        clusterReducer(
          {
            ...initialTestState,
            clusters: {
              default: {
                currentNamespace: "",
                namespaces: ["default"],
                canCreateNS: true,
              },
              other: {
                currentNamespace: "",
                namespaces: ["othernamespace"],
                error: { action: "create", error: new Error("boom") },
                canCreateNS: true,
              },
            },
          } as IClustersState,
          {
            type: getType(actions.namespace.receiveNamespaces),
            payload: {
              cluster: "other",
              namespaces: ["one", "two", "three"],
            },
          },
        ),
      ).toEqual({
        ...initialTestState,
        clusters: {
          default: {
            currentNamespace: "",
            canCreateNS: true,
            namespaces: ["default"],
          },
          other: {
            currentNamespace: "one",
            namespaces: ["one", "two", "three"],
            error: undefined,
            canCreateNS: true,
          },
        },
      } as IClustersState);
    });

    it("gets the namespace from the token", () => {
      Auth.defaultNamespaceFromToken = jest.fn(() => "two");
      expect(
        clusterReducer(
          {
            ...initialTestState,
            clusters: {
              other: {
                currentNamespace: "",
                namespaces: [],
                canCreateNS: true,
              },
            },
          } as IClustersState,
          {
            type: getType(actions.namespace.receiveNamespaces),
            payload: {
              cluster: "other",
              namespaces: ["one", "two", "three"],
            },
          },
        ),
      ).toEqual({
        ...initialTestState,
        clusters: {
          other: {
            currentNamespace: "two",
            namespaces: ["one", "two", "three"],
            error: undefined,
            canCreateNS: true,
          },
        },
      } as IClustersState);
    });

    it("defaults to the first namespace if the one in token is not available", () => {
      Auth.defaultNamespaceFromToken = jest.fn(() => "nope");
      expect(
        clusterReducer(
          {
            ...initialTestState,
            clusters: {
              other: {
                currentNamespace: "",
                namespaces: [],
                canCreateNS: true,
              },
            },
          } as IClustersState,
          {
            type: getType(actions.namespace.receiveNamespaces),
            payload: {
              cluster: "other",
              namespaces: ["one", "two", "three"],
            },
          },
        ),
      ).toEqual({
        ...initialTestState,
        clusters: {
          other: {
            currentNamespace: "one",
            namespaces: ["one", "two", "three"],
            error: undefined,
            canCreateNS: true,
          },
        },
      } as IClustersState);
    });

    it("gets the existing current namespace", () => {
      expect(
        clusterReducer(
          {
            ...initialTestState,
            clusters: {
              other: {
                currentNamespace: "three",
                namespaces: [],
                canCreateNS: true,
              },
            },
          } as IClustersState,
          {
            type: getType(actions.namespace.receiveNamespaces),
            payload: {
              cluster: "other",
              namespaces: ["one", "two", "three"],
            },
          },
        ),
      ).toEqual({
        ...initialTestState,
        clusters: {
          other: {
            currentNamespace: "three",
            namespaces: ["one", "two", "three"],
            error: undefined,
            canCreateNS: true,
          },
        },
      } as IClustersState);
    });

    it("gets the stored namespace", () => {
      jest
        .spyOn(window.localStorage.__proto__, "getItem")
        .mockReturnValueOnce('{"other": "three"}');
      expect(
        clusterReducer(
          {
            ...initialTestState,
            clusters: {
              other: {
                currentNamespace: "",
                namespaces: [],
                canCreateNS: true,
              },
            },
          } as IClustersState,
          {
            type: getType(actions.namespace.receiveNamespaces),
            payload: {
              cluster: "other",
              namespaces: ["one", "two", "three"],
            },
          },
        ),
      ).toEqual({
        ...initialTestState,
        clusters: {
          other: {
            currentNamespace: "three",
            namespaces: ["one", "two", "three"],
            error: undefined,
            canCreateNS: true,
          },
        },
      } as IClustersState);
    });

    it("ignores the stored namespace if it's not available", () => {
      jest.spyOn(window.localStorage.__proto__, "getItem").mockReturnValueOnce('{"other": "four"}');
      expect(
        clusterReducer(
          {
            ...initialTestState,
            clusters: {
              other: {
                currentNamespace: "",
                namespaces: [],
                canCreateNS: true,
              },
            },
          } as IClustersState,
          {
            type: getType(actions.namespace.receiveNamespaces),
            payload: {
              cluster: "other",
              namespaces: ["one", "two", "three"],
            },
          },
        ),
      ).toEqual({
        ...initialTestState,
        clusters: {
          other: {
            currentNamespace: "one",
            namespaces: ["one", "two", "three"],
            error: undefined,
            canCreateNS: true,
          },
        },
      } as IClustersState);
    });
  });

  context("when RECEIVE_CONFIG", () => {
    const config = {
      kubeappsCluster: "",
      kubeappsNamespace: "kubeapps",
      helmGlobalNamespace: "kubeapps-global",
      carvelGlobalNamespace: "kapp-controller-packaging-global",
      appVersion: "dev",
      authProxyEnabled: false,
      oauthLoginURI: "",
      oauthLogoutURI: "",
      featureFlags: {
        operators: false,
        ui: "hex",
      },
      clusters: ["additionalCluster1", "additionalCluster2"],
      authProxySkipLoginPage: false,
      theme: "light",
      remoteComponentsUrl: "",
      customAppViews: [],
      skipAvailablePackageDetails: false,
      createNamespaceLabels: {},
      configuredPlugins: [],
    } as IConfig;
    it("re-writes the clusters to match the config.clusters state", () => {
      expect(
        clusterReducer(initialTestState, {
          type: getType(actions.config.receiveConfig),
          payload: config,
        }),
      ).toEqual({
        ...initialTestState,
        currentCluster: "additionalCluster1",
        clusters: {
          additionalCluster1: {
            currentNamespace: "",
            namespaces: [],
            canCreateNS: false,
          },
          additionalCluster2: {
            currentNamespace: "",
            namespaces: [],
            canCreateNS: false,
          },
        },
      } as IClustersState);
    });

    it("sets the current cluster to the first cluster in the list", () => {
      const configClusters = {
        ...config,
        clusters: ["one", "two", "three"],
      };
      expect(
        clusterReducer(initialTestState, {
          type: getType(actions.config.receiveConfig),
          payload: configClusters,
        }),
      ).toEqual({
        ...initialTestState,
        currentCluster: "one",
        clusters: {
          one: {
            currentNamespace: "",
            namespaces: [],
            canCreateNS: false,
          },
          two: {
            currentNamespace: "",
            namespaces: [],
            canCreateNS: false,
          },
          three: {
            currentNamespace: "",
            namespaces: [],
            canCreateNS: false,
          },
        },
      } as IClustersState);
    });
  });
  context("when ALLOW_CREATE_NAMESPACE", () => {
    afterEach(() => {
      jest.restoreAllMocks();
    });

    it("allows to create ns", () => {
      expect(
        clusterReducer(
          {
            ...initialTestState,
            clusters: {
              default: {
                currentNamespace: "",
                namespaces: ["default"],
                canCreateNS: false,
              },
            },
          } as IClustersState,
          {
            type: getType(actions.namespace.setAllowCreate),
            payload: {
              cluster: "default",
              allowed: true,
            },
          },
        ),
      ).toEqual({
        ...initialTestState,
        clusters: {
          default: {
            currentNamespace: "",
            canCreateNS: true,
            namespaces: ["default"],
          },
        },
      } as IClustersState);
    });
  });
});
