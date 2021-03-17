import { LOCATION_CHANGE, LocationChangeAction } from "connected-react-router";

import { getType } from "typesafe-actions";
import actions from "../actions";
import { KubeAction } from "../actions/kube";
import { IK8sList, IKubeState, IResource } from "../shared/types";

export const initialKinds = {
  // In case it's not possible to retrieve the api groups (e.g. with lacking permissions)
  // Use pre-defined settings. Obtained with Kubernetes v1.17
  APIService: { apiVersion: "apiregistration.k8s.io/v1", plural: "apiservices", namespaced: false },
  AppRepository: {
    apiVersion: "kubeapps.com/v1alpha1",
    plural: "apprepositories",
    namespaced: true,
  },
  Binding: { apiVersion: "v1", plural: "bindings", namespaced: true },
  CSINode: { apiVersion: "storage.k8s.io/v1", plural: "csinodes", namespaced: false },
  ClusterRole: {
    apiVersion: "rbac.authorization.k8s.io/v1",
    plural: "clusterroles",
    namespaced: false,
  },
  ClusterRoleBinding: {
    apiVersion: "rbac.authorization.k8s.io/v1",
    plural: "clusterrolebindings",
    namespaced: false,
  },
  ComponentStatus: { apiVersion: "v1", plural: "componentstatuses", namespaced: false },
  ConfigMap: { apiVersion: "v1", plural: "configmaps", namespaced: true },
  ControllerRevision: { apiVersion: "apps/v1", plural: "controllerrevisions", namespaced: true },
  CustomResourceDefinition: {
    apiVersion: "apiextensions.k8s.io/v1",
    plural: "customresourcedefinitions",
    namespaced: false,
  },
  DaemonSet: { apiVersion: "apps/v1", plural: "daemonsets", namespaced: true },
  Deployment: { apiVersion: "apps/v1", plural: "deployments", namespaced: true },
  Endpoints: { apiVersion: "v1", plural: "endpoints", namespaced: true },
  HorizontalPodAutoscaler: {
    apiVersion: "autoscaling/v1",
    plural: "horizontalpodautoscalers",
    namespaced: true,
  },
  Ingress: { apiVersion: "extensions/v1beta1", plural: "ingresses", namespaced: true },
  Job: { apiVersion: "batch/v1", plural: "jobs", namespaced: true },
  Lease: { apiVersion: "coordination.k8s.io/v1", plural: "leases", namespaced: true },
  LimitRange: { apiVersion: "v1", plural: "limitranges", namespaced: true },
  LocalSubjectAccessReview: {
    apiVersion: "authorization.k8s.io/v1",
    plural: "localsubjectaccessreviews",
    namespaced: true,
  },
  MutatingWebhookConfiguration: {
    apiVersion: "admissionregistration.k8s.io/v1",
    plural: "mutatingwebhookconfigurations",
    namespaced: false,
  },
  Namespace: { apiVersion: "v1", plural: "namespaces", namespaced: false },
  NetworkPolicy: {
    apiVersion: "networking.k8s.io/v1",
    plural: "networkpolicies",
    namespaced: true,
  },
  Node: { apiVersion: "v1", plural: "nodes", namespaced: false },
  PersistentVolume: { apiVersion: "v1", plural: "persistentvolumes", namespaced: false },
  PersistentVolumeClaim: { apiVersion: "v1", plural: "persistentvolumeclaims", namespaced: true },
  Pod: { apiVersion: "v1", plural: "pods", namespaced: true },
  PodTemplate: { apiVersion: "v1", plural: "podtemplates", namespaced: true },
  PriorityClass: {
    apiVersion: "scheduling.k8s.io/v1",
    plural: "priorityclasses",
    namespaced: false,
  },
  ReplicaSet: { apiVersion: "apps/v1", plural: "replicasets", namespaced: true },
  ReplicationController: { apiVersion: "v1", plural: "replicationcontrollers", namespaced: true },
  ResourceQuota: { apiVersion: "v1", plural: "resourcequotas", namespaced: true },
  Role: { apiVersion: "rbac.authorization.k8s.io/v1", plural: "roles", namespaced: true },
  RoleBinding: {
    apiVersion: "rbac.authorization.k8s.io/v1",
    plural: "rolebindings",
    namespaced: true,
  },
  Secret: { apiVersion: "v1", plural: "secrets", namespaced: true },
  SelfSubjectAccessReview: {
    apiVersion: "authorization.k8s.io/v1",
    plural: "selfsubjectaccessreviews",
    namespaced: false,
  },
  SelfSubjectRulesReview: {
    apiVersion: "authorization.k8s.io/v1",
    plural: "selfsubjectrulesreviews",
    namespaced: false,
  },
  Service: { apiVersion: "v1", plural: "services", namespaced: true },
  ServiceAccount: { apiVersion: "v1", plural: "serviceaccounts", namespaced: true },
  StatefulSet: { apiVersion: "apps/v1", plural: "statefulsets", namespaced: true },
  StorageClass: { apiVersion: "storage.k8s.io/v1", plural: "storageclasses", namespaced: false },
  SubjectAccessReview: {
    apiVersion: "authorization.k8s.io/v1",
    plural: "subjectaccessreviews",
    namespaced: false,
  },
  TokenReview: {
    apiVersion: "authentication.k8s.io/v1",
    plural: "tokenreviews",
    namespaced: false,
  },
  ValidatingWebhookConfiguration: {
    apiVersion: "admissionregistration.k8s.io/v1",
    plural: "validatingwebhookconfigurations",
    namespaced: false,
  },
  VolumeAttachment: {
    apiVersion: "storage.k8s.io/v1",
    plural: "volumeattachments",
    namespaced: false,
  },
};

export const initialState: IKubeState = {
  items: {},
  kinds: initialKinds,
  sockets: {},
  timers: {},
};

const kubeReducer = (
  state: IKubeState = initialState,
  action: KubeAction | LocationChangeAction,
): IKubeState => {
  let key: string;
  switch (action.type) {
    case getType(actions.kube.requestResource): {
      let item = state.items[action.payload];
      if (!item) {
        item = { isFetching: true };
      }
      const requestedItem = { [action.payload]: item };
      return { ...state, items: { ...state.items, ...requestedItem } };
    }
    case getType(actions.kube.receiveResource): {
      const receivedItem = {
        [action.payload.key]: { isFetching: false, item: action.payload.resource },
      };
      return { ...state, items: { ...state.items, ...receivedItem } };
    }
    case getType(actions.kube.receiveResourceKinds):
      return { ...state, kinds: action.payload };
    case getType(actions.kube.receiveKindsError):
      return { ...state, kinds: initialKinds, kindsError: action.payload };
    case getType(actions.kube.receiveResourceFromList): {
      const stateListItem = state.items[action.payload.key].item as IK8sList<IResource, {}>;
      const newItem = action.payload.resource as IResource;
      if (!stateListItem || !stateListItem.items) {
        return {
          ...state,
          items: {
            ...state.items,
            [action.payload.key]: {
              isFetching: false,
              item: { ...stateListItem, items: [newItem] },
            },
          },
        };
      }
      const updatedItems = stateListItem.items.map(it => {
        if (it.metadata.selfLink === newItem.metadata.selfLink) {
          return action.payload.resource as IResource;
        }
        return it;
      });
      return {
        ...state,
        items: {
          ...state.items,
          [action.payload.key]: {
            isFetching: false,
            item: { ...stateListItem, items: updatedItems },
          },
        },
      };
    }
    case getType(actions.kube.receiveResourceError): {
      const erroredItem = {
        [action.payload.key]: { isFetching: false, error: action.payload.error },
      };
      return { ...state, items: { ...state.items, ...erroredItem } };
    }
    case getType(actions.kube.openWatchResource): {
      const { ref, handler, onError } = action.payload;
      key = ref.watchResourceURL();
      if (state.sockets[key]) {
        // Socket for this resource already open, do nothing
        return state;
      }
      const socket = ref.watchResource();
      socket.addEventListener("message", handler);
      socket.addEventListener("error", onError);
      return {
        ...state,
        sockets: {
          ...state.sockets,
          [key]: { socket, onError },
        },
      };
    }
    // TODO(adnan): this won't handle cases where one component closes a socket
    // another one is using. Whilst not a problem today, a reference counter
    // approach could be used here to enable this in the future.
    case getType(actions.kube.closeWatchResource): {
      key = action.payload.watchResourceURL();
      const { sockets } = state;
      const { [key]: foundSocket, ...otherSockets } = sockets;
      const timerID = action.payload.getResourceURL();
      const timer = state.timers[timerID];
      // close the socket if it exists
      if (foundSocket !== undefined) {
        foundSocket.socket.removeEventListener("error", foundSocket.onError);
        foundSocket.socket.close();
      }
      if (timer) {
        clearInterval(timer);
      }
      return {
        ...state,
        sockets: otherSockets,
        timers: {
          ...state.timers,
          [timerID]: undefined,
        },
      };
    }
    case getType(actions.kube.addTimer): {
      if (!state.timers[action.payload.id]) {
        return {
          ...state,
          timers: {
            ...state.timers,
            [action.payload.id]: setInterval(action.payload.timer, 5000),
          },
        };
      }
      return state;
    }
    case getType(actions.kube.removeTimer): {
      if (state.timers[action.payload]) {
        clearInterval(state.timers[action.payload] as NodeJS.Timer);
        return {
          ...state,
          timers: {
            ...state.timers,
            [action.payload]: undefined,
          },
        };
      }
      return state;
    }
    case LOCATION_CHANGE:
      return { ...state, items: {} };
    default:
  }
  return state;
};

export default kubeReducer;
