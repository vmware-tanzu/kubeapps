import { LOCATION_CHANGE, LocationChangeAction } from "connected-react-router";
import { getType } from "typesafe-actions";

import actions from "../actions";
import { AuthAction } from "../actions/auth";
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
      "default": {
        currentNamespace: Auth.defaultNamespaceFromToken(token),
        namespaces: [],
      },
    },
  } as IClustersState;
};
const initialState: IClustersState = getInitialState();

const clusterReducer = (
  state: IClustersState = initialState,
  action: NamespaceAction | LocationChangeAction | AuthAction,
): IClustersState => {
  switch (action.type) {
    case getType(actions.namespace.receiveNamespace):
      if (!state.clusters.default.namespaces.includes(action.payload.metadata.name)) {
        return {
          ...state,
          clusters: {
            default: {
              ...state.clusters.default,
              namespaces: state.clusters.default.namespaces.concat(action.payload.metadata.name).sort(),
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
          default: {
            ...state.clusters.default,
            namespaces: action.payload,
          },
        },
      };
    case getType(actions.namespace.setNamespace):
      return {
        ...state,
        clusters: {
          default: {
            ...state.clusters.default,
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
          default: {
            ...state.clusters.default,
            error: { action: action.payload.op, error: action.payload.err },
          },
        },
      };
    case getType(actions.namespace.clearNamespaces):
      // TODO(absoludity): this should maintain the keys for all clusters.
      return { ...initialState };
    case LOCATION_CHANGE:
      const pathname = action.payload.location.pathname;
      // looks for /ns/:namespace in URL
      // TODO(absoludity): this should match on cluster also to set currentCluster.
      const matches = pathname.match(/\/ns\/([^/]*)/);
      if (matches) {
        return {
          ...state,
          clusters: {
            default: {
              ...state.clusters.default,
              currentNamespace: matches[1]
            },
          },
        };
      }
      break;
    case getType(actions.auth.setAuthenticated):
      // Only when a user is authenticated to we set the current namespace from
      // the auth default namespace.
      if (action.payload.authenticated) {
        return {
          ...state,
          clusters: {
            default: {
              ...state.clusters.default,
              currentNamespace: action.payload.defaultNamespace,
            },
          },
        };
      }
    default:
  }
  return state;
};

export default clusterReducer;
