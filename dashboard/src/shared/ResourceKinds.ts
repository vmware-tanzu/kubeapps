// We explicitly define the plurals here, just in case a generic pluralizer
// isn't sufficient. Note that CRDs can explicitly define pluralized forms,
// which might not match with the Kind. If this becomes difficult to
// maintain we can add a generic pluralizer and a way to override.
export const ResourceKindsWithPlurals = {
  ClusterRole: "clusterroles",
  ConfigMap: "configmaps",
  DaemonSet: "daemonsets",
  Deployment: "deployments",
  Ingress: "ingresses",
  Secret: "secrets",
  Service: "services",
  StatefulSet: "statefulsets",
} as const;

export type ResourceKind = keyof typeof ResourceKindsWithPlurals;
