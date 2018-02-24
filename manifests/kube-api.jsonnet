local kube = import "kube.libsonnet";

{
  namespace:: {metadata+: {namespace: "kubeapps"}},

  serviceAccount: kube.ServiceAccount("kubeapps-kube-api") + $.namespace,

  role: kube.Role("kubeapps-kube-api") + $.namespace {
    rules: [
      // Kubeapps reads Helm release data from the ConfigMaps Tiller store.
      {
        apiGroups: [""],
        resources: ["configmaps"],
        verbs: ["get", "list"],
      },
      // Kubeapps creates and manages AppRepository CRD objects that define
      // which application (e.g. chart) repositories will be indexed.
      {
        apiGroups: ["kubeapps.com"],
        resources: ["apprepositories"],
        verbs: ["get", "list", "create", "update", "delete"],
      },
    ],
  },
  roleBinding: kube.RoleBinding("kubeapps-kube-api") + $.namespace {
    roleRef_: $.role,
    subjects_: [$.serviceAccount],
  },

  clusterRole: kube.ClusterRole("kubeapps-kube-api") {
    rules: [
      // Kubeapps creates and manages Helm releases via the HelmRelease CRD
      // object. See https://github.com/bitnami-labs/helm-crd for more info.
      {
        apiGroups: ["helm.bitnami.com"],
        resources: ["helmreleases"],
        verbs: ["get", "list", "create", "update", "delete"],
      },
      // Kubeapps watches Deployments and Services to monitor updates for apps
      // in the UI.
      {
        apiGroups: ["", "apps"],
        resources: ["services", "deployments"],
        verbs: ["list", "watch"],
      },
      // Kubeapps lists available Service Brokers and can request relisting
      // using patch.
      {
        apiGroups: ["servicecatalog.k8s.io"],
        resources: ["clusterservicebrokers"],
        verbs: ["list", "patch"],
      },
      // Kubeapps lists available Service Classes and Plans.
      {
        apiGroups: ["servicecatalog.k8s.io"],
        resources: ["clusterserviceclasses", "clusterserviceplans"],
        verbs: ["list"],
      },
      // Kubeapps creates and manages Service Catalog Instances and Bindings.
      {
        apiGroups: ["servicecatalog.k8s.io"],
        resources: ["serviceinstances", "servicebindings"],
        verbs: ["get", "list", "create", "delete"],
      },
      // Kubeapps displays Secrets from Service Bindings, which could be in any
      // namespace.
      {
        apiGroups: [""],
        resources: ["secrets"],
        verbs: ["get"],
      },
    ]
  },
  clusterRoleBinding: kube.ClusterRoleBinding("kubeapps-kube-api") {
    roleRef_: $.clusterRole,
    subjects_: [$.serviceAccount],
  },

  service: kube.Service("kubeapps-kube-api") + $.namespace {
    target_pod: $.deployment.spec.template,
  },

  deployment: kube.Deployment("kubeapps-kube-api") + $.namespace {
    spec+: {
      template+: {
        spec+: {
          serviceAccountName: $.serviceAccount.metadata.name,
          containers_+: {
            default: kube.Container("proxy") {
              image: "lachlanevenson/k8s-kubectl:v1.8.6",
              args: [
                "proxy",
                "--address=0.0.0.0",
                "--port=8080"
              ],
              ports_: {
                http: {containerPort: 8080, protocol: "TCP"},
              },
            },
          },
        },
      },
    },
  },
}
