// tiller with a proxy running as a sidecar

local kube = import "kube.libsonnet";

// Run Proxy as a sidecar, and restrict tiller port to pod-only

{
  namespace:: {metadata+: {namespace: "kube-system"}},
  local proxyContainer = kube.Container("proxy") {
    image: "kubeapps/tiller-proxy:" + std.extVar("VERSION"),
    securityContext: {
      readOnlyRootFilesystem: true,
    },
    ports_: {
      http: {containerPort: 8080}
    },
    command: ["/proxy"],
    args_: {
      host: "localhost:44134",
    },
    env_: {
      POD_NAMESPACE: $.namespace.metadata.namespace,
    },
  },
  tillerServiceAccount: kube.ServiceAccount("tiller") + $.namespace,
  tillerBinding: kube.ClusterRoleBinding("tiller-cluster-admin") {
    roleRef_: kube.ClusterRole("cluster-admin"),
    subjects_: [$.tillerServiceAccount],
  },
  local tillerDeploy = (import "tiller-deployment.jsonnet"),
  local tillerContainer = tillerDeploy.spec.template.spec.containers[0],
  tillerProxy: tillerDeploy + $.namespace {
    spec+: {
      template+: {
        spec+: {
          serviceAccountName: $.tillerServiceAccount.metadata.name,
          containers: [proxyContainer, tillerContainer {
            command: ["/tiller"],
            args+: ["--listen=localhost:44134"],  // remove access to :44134 outside pod
            overrideEnvs(overrides):: [
              if std.objectHas(overrides, x.name) then { name: x.name, value: overrides[x.name] } else x for x in tillerContainer.env
            ],
            env: self.overrideEnvs({
              TILLER_NAMESPACE: $.namespace.metadata.namespace,
            }),
          }],
        },
      },
    },
  },

  service: kube.Service("tiller-deploy") + $.namespace {
    target_pod: $.tillerProxy.spec.template,
  }
}
