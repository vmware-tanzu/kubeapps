local kube = import "kube.libsonnet";

{
  namespace:: {metadata+: {namespace: "kubeapps"}},

  serviceAccount: kube.ServiceAccount("kubeapps-kube-api") + $.namespace,
  // TODO: create restricted set of roles for this API proxy
  binding: kube.ClusterRoleBinding("kubeapps-kube-api") {
    roleRef_: kube.ClusterRole("cluster-admin"),
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
