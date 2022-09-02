// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { LocationChangeAction, LOCATION_CHANGE } from "connected-react-router";
import { IKubeState } from "shared/types";
import { Kube } from "shared/Kube";
import { getType } from "typesafe-actions";
import actions from "../actions";
import { KubeAction } from "../actions/kube";

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
  Ingress: { apiVersion: "extensions/v1", plural: "ingresses", namespaced: true },
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
  // We book keep on subscriptions, keyed by the installed package ref,
  // so that we can unsubscribe when the closeRequestResources action is
  // dispatched (usually because the component is unmounted when the user
  // navigates away).
  subscriptions: {},
};

const kubeReducer = (
  state: IKubeState = initialState,
  action: KubeAction | LocationChangeAction,
): IKubeState => {
  switch (action.type) {
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
    case getType(actions.kube.receiveResourceError): {
      const erroredItem = {
        [action.payload.key]: { isFetching: false, error: action.payload.error },
      };
      return { ...state, items: { ...state.items, ...erroredItem } };
    }
    case getType(actions.kube.requestResources): {
      const { pkg, refs, handler, watch, onError, onComplete } = action.payload;
      const observable = Kube.getResources(pkg, refs, watch);
      const subscription = observable.subscribe({
        next(r) {
          handler(r);
        },
        error(e) {
          onError(e);
        },
        complete() {
          onComplete();
        },
      });
      // We only record the subscription if watching the result, since otherwise
      // the call is terminated by the server automatically once results are
      // returned and we don't need any book-keeping.
      if (watch) {
        const key = `${pkg.context?.cluster}/${pkg.context?.namespace}/${pkg.identifier}`;
        return {
          ...state,
          subscriptions: {
            ...state.subscriptions,
            [key]: subscription,
          },
        };
      }
      return state;
    }
    case getType(actions.kube.closeRequestResources): {
      const pkg = action.payload;
      const key = `${pkg.context?.cluster}/${pkg.context?.namespace}/${pkg.identifier}`;
      const { subscriptions } = state;
      const { [key]: foundSubscription, ...otherSubscriptions } = subscriptions;
      // unsubscribe if it exists
      if (foundSubscription !== undefined) {
        foundSubscription.unsubscribe();
      }
      return {
        ...state,
        subscriptions: otherSubscriptions,
      };
    }
    case LOCATION_CHANGE:
      return { ...state, items: {} };
    default:
  }
  return state;
};

export default kubeReducer;
