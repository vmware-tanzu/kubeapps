import { LOCATION_CHANGE, LocationChangeAction } from "connected-react-router";
import { getType } from "typesafe-actions";

import { IConfig } from "shared/Config";
import { definedNamespaces } from "shared/Namespace";
import actions from "../actions";
import { AuthAction } from "../actions/auth";
import { ConfigAction } from "../actions/config";
import { NamespaceAction } from "../actions/namespace";
import { Auth } from "../shared/Auth";

export interface IClusterState {
  currentNamespace: string;
  namespaces: string[];
  error?: { action: string; error: Error };
}

interface IClustersMap {
  [cluster: string]: IClusterState;
}

export interface IClustersState {
  currentCluster: string;
  clusters: IClustersMap;
}

const getInitialState: () => IClustersState = (): IClustersState => {
  const token = Auth.getAuthToken() || "";
  return {
    currentCluster: "default",
    clusters: {
      default: {
        currentNamespace: Auth.defaultNamespaceFromToken(token),
        namespaces: [],
      },
    },
  } as IClustersState;
};
export const initialState: IClustersState = getInitialState();

const clusterReducer = (
  state: IClustersState = initialState,
  action: ConfigAction | NamespaceAction | LocationChangeAction | AuthAction,
): IClustersState => {
  switch (action.type) {
    case getType(actions.namespace.receiveNamespace):
      if (!state.clusters.default.namespaces.includes(action.payload.namespace.metadata.name)) {
        return {
          ...state,
          clusters: {
            ...state.clusters,
            [action.payload.cluster]: {
              ...state.clusters[action.payload.cluster],
              namespaces: state.clusters[action.payload.cluster].namespaces
                .concat(action.payload.namespace.metadata.name)
                .sort(),
              error: undefined,
            },
          },
        };
      }
      return state;
    case getType(actions.namespace.receiveNamespaces):
      return {
        ...state,
        clusters: {
          ...state.clusters,
          [action.payload.cluster]: {
            ...state.clusters[action.payload.cluster],
            namespaces: action.payload.namespaces,
            error: undefined,
          },
        },
      };
    case getType(actions.namespace.setNamespace):
      return {
        ...state,
        clusters: {
          ...state.clusters,
          [state.currentCluster]: {
            ...state.clusters[state.currentCluster],
            currentNamespace: action.payload,
            error: undefined,
          },
        },
      };
    case getType(actions.namespace.errorNamespaces):
      // Ignore error listing namespaces since those are expected
      if (action.payload.op === "list") {
        return state;
      }
      return {
        ...state,
        clusters: {
          ...state.clusters,
          [action.payload.cluster]: {
            ...state.clusters[action.payload.cluster],
            error: { action: action.payload.op, error: action.payload.err },
          },
        },
      };
    case getType(actions.namespace.clearClusters):
      return {
        ...state,
        clusters: {
          ...initialState.clusters,
        },
      };
    case LOCATION_CHANGE:
      const pathname = action.payload.location.pathname;
      // looks for either or both of /c/:cluster and /ns/:namespace in URL
      const matches = pathname.match(/(?:\/c\/(?<cluster>[^/]*))?(?:\/ns\/(?<namespace>[^/]*))?/);
      if (matches && matches.groups) {
        let [currentCluster, currentNamespace] = [matches.groups.cluster, matches.groups.namespace];
        currentCluster = currentCluster || state.currentCluster;
        currentNamespace = currentNamespace || state.clusters[currentCluster].currentNamespace;
        return {
          ...state,
          currentCluster,
          clusters: {
            ...state.clusters,
            [currentCluster]: {
              ...state.clusters[currentCluster],
              currentNamespace,
            },
          },
        };
      }
      break;
    case getType(actions.auth.setAuthenticated):
      // Only when a user is authenticated to we set the current namespace from
      // the auth default namespace.
      if (action.payload.authenticated) {
        if (state.clusters[state.currentCluster].currentNamespace === definedNamespaces.all) {
          return {
            ...state,
            clusters: {
              ...state.clusters,
              [state.currentCluster]: {
                ...state.clusters[state.currentCluster],
                currentNamespace: action.payload.defaultNamespace,
              },
            },
          };
        }
      }
      break;
    case getType(actions.config.receiveConfig):
      // Initialize the additional clusters when receiving the config.
      const clusters = {
        ...state.clusters,
      };
      const config = action.payload as IConfig;
      config.clusters?.forEach(cluster => {
        clusters[cluster.name] = {
          currentNamespace: "default",
          namespaces: [],
        };
      });
      return {
        ...state,
        clusters,
      };
    default:
  }
  return state;
};

export default clusterReducer;
