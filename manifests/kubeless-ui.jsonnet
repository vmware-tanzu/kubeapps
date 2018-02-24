local kube = import "kube.libsonnet";

local host = "kubeless-ui";

{
  namespace:: {metadata+: {namespace: "kubeless"}},

  serviceAccount: kube.ServiceAccount("kubeless-ui") + $.namespace,

  editorRole: kube.ClusterRole("kubeless-editor") {
    rules: [
      {
        apiGroups: ["kubeless.io"],
        resources: ["functions"],
        verbs: ["get", "list", "watch", "create", "patch", "delete"],
      },
      {
        apiGroups: [""],
        resources: ["pods","pods/log"],
        verbs: ["get", "list"],
      },
      {
        apiGroups: [""],
        resources: ["services","services/proxy"],
        verbs: ["get", "list", "proxy"],
      },
    ],
  },

  editorBinding: kube.ClusterRoleBinding("kubeless-ui-editor") {
    roleRef_: $.editorRole,
    subjects_: [$.serviceAccount],
  },

  svc: kube.Service("kubeless-ui") + $.namespace {
    target_pod: $.deploy.spec.template,
  },

  deploy: kube.Deployment("kubeless-ui") + $.namespace {
    spec+: {
      template+: {
        spec+: {
          serviceAccountName: $.serviceAccount.metadata.name,
          containers_+: {
            default: kube.Container("ui") {
              image: "bitnami/kubeless-ui@sha256:8784735d3248f9ede6f869c1e0e6b39cb16c9738411fe60e3ce54584f6e51c20",
              ports_: {
                ui: {containerPort: 3000, protocol: "TCP"},
              },
              readinessProbe: {
                httpGet: {path: "/", port: 3000},
              },
            },
            proxy: kube.Container("proxy") {
              image: "kelseyhightower/kubectl:1.4.0",
              args: ["proxy", "-p", "8080"],
            },
          },
        },
      },
    },
  },
}
