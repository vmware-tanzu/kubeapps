// We explicitly define the apiVersions here, just in case a generic pluralizer
// isn't sufficient. Note that apiVersions may change over time.
export const ResourceKindsWithAPIVersions = {
  ClusterRole: "rbac.authorization.k8s.io/v1",
  ClusterRoleBinding: "rbac.authorization.k8s.io/v1",
  ConfigMap: "v1",
  DaemonSet: "apps/v1",
  Deployment: "apps/v1",
  Ingress: "extensions/v1beta1",
  ReplicaSet: "apps/v1",
  Role: "rbac.authorization.k8s.io/v1",
  RoleBinding: "rbac.authorization.k8s.io/v1",
  Secret: "v1",
  Service: "v1",
  ServiceAccount: "v1",
  StatefulSet: "apps/v1",
  PersistentVolumeClaim: "v1",
  Pod: "v1",
} as const;

export type ResourceAPIVersion = keyof typeof ResourceKindsWithAPIVersions;
