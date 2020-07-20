// We explicitly define the plurals here, just in case a generic pluralizer
// isn't sufficient. Note that CRDs can explicitly define pluralized forms,
// which might not match with the Kind. If this becomes difficult to
// maintain we can add a generic pluralizer and a way to override.
export const ResourceKindsWithPlurals = {
  ClusterRole: "clusterroles",
  ClusterRoleBinding: "clusterrolebindings",
  ConfigMap: "configmaps",
  DaemonSet: "daemonsets",
  Deployment: "deployments",
  Ingress: "ingresses",
  ReplicaSet: "replicasets",
  Role: "roles",
  RoleBinding: "rolebindings",
  Secret: "secrets",
  Service: "services",
  ServiceAccount: "serviceaccounts",
  StatefulSet: "statefulsets",
  PersistentVolumeClaim: "persistentvolumeclaims",
  Pod: "pods",
} as const;

export const isNamespaced = (resource: string) => {
  switch (resource) {
    case "clusterroles":
    case "clusterrolebindings":
      return false;
    default:
      return true;
  }
};

export type ResourceKind = keyof typeof ResourceKindsWithPlurals;
