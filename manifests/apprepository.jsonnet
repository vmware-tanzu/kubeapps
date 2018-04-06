local kube = import "kube.libsonnet";
local kubecfg = import "kubecfg.libsonnet";

local labels = {
  app: "apprepository-controller",
};

{
  namespace:: {metadata+: {namespace: "kubeapps"}},

  crd: kube.CustomResourceDefinition("apprepositories.kubeapps.com") {
    spec+: {
      group: "kubeapps.com",
      version: "v1alpha1",
      names: {
        kind: "AppRepository",
        plural: "apprepositories",
        shortNames: ["apprepos"],
      },
    },
  },

  serviceaccount: kube.ServiceAccount("apprepository-controller") + $.namespace,

  // Need a cluster role because client-go v5.0.1 does not support namespaced informers
  // TODO: remove when we update to client-go v6.0.0
  clusterRole: kube.ClusterRole("apprepository-controller") {
    rules: [
      {
        apiGroups: ["batch"],
        resources: ["cronjobs"],
        verbs: ["get", "list", "watch"],
      },
    ]
  },
  clusterRoleBinding: kube.ClusterRoleBinding("apprepository-controller") {
    roleRef_: $.clusterRole,
    subjects_: [$.serviceaccount],
  },

  role: kube.Role("apprepository-controller") + $.namespace {
    rules: [
      {
        apiGroups: [""],
        resources: ["events"],
        verbs: ["create"],
      },
      {
        apiGroups: ["batch"],
        resources: ["cronjobs"],
        verbs: ["create", "get", "list", "update", "watch"],
      },
      {
        apiGroups: ["batch"],
        resources: ["jobs"],
        verbs: ["create"],
      },
      {
        apiGroups: ["kubeapps.com"],
        resources: ["apprepositories"],
        verbs: ["get", "list", "update", "watch"],
      },
    ]
  },

  rolebinding: kube.RoleBinding("apprepository-controller") + $.namespace {
    roleRef_: $.role,
    subjects_: [$.serviceaccount],
  },

  deployment: kube.Deployment("apprepository-controller") + $.namespace {
    metadata+: {labels+: labels},
    spec+: {
      template+: {
        spec+: {
          serviceAccountName: $.serviceaccount.metadata.name,
          containers_+: {
            default: kube.Container("controller") {
              image: "kubeapps/apprepository-controller:" + std.extVar("VERSION"),
              command: ["/apprepository-controller"],
              args: ["--logtostderr", "--repo-sync-image=kubeapps/chart-repo:" + std.extVar("VERSION")],
            },
          },
        },
      },
    },
  },

  _apprepo(name, url):: kube._Object("kubeapps.com/v1alpha1", "AppRepository", name) + $.namespace {
    spec: {
      url: url,
      type: "helm"
    },
  },

  apprepos: {
    stable: $._apprepo("stable", "https://kubernetes-charts.storage.googleapis.com"),
    incubator: $._apprepo("incubator", "https://kubernetes-charts-incubator.storage.googleapis.com"),
    svccat: $._apprepo("svc-cat", "https://svc-catalog-charts.storage.googleapis.com"),
  },
}
