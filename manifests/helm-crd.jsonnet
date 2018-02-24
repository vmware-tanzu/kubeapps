// CRD and tiller, with controller running as a sidecar

local kube = import "kube.libsonnet";

// Run CRD controller as a sidecar, and restrict tiller port to pod-only
local controllerOverlay = {
  spec+: {
    template+: {
      spec+: {
        volumes+: [
          // Used as temporary space while downloading charts, etc.
          {name: "home", emptyDir: {}},
        ],
        containers+: [
          kube.Container("controller") {
            name: "controller",
            image: "bitnami/helm-crd-controller:v0.2.1",
            securityContext: {
              readOnlyRootFilesystem: true,
            },
            command: ["/controller"],
            args_: {
              home: "/helm",
              host: "localhost:44134",
            },
            env_: {
              TMPDIR: "/helm",
            },
            volumeMounts_: {
              home: {mountPath: "/helm"},
            },
          },
        ],
      },
    },
  },
};

{
  namespace:: {metadata+: {namespace: "kube-system"}},
  crd: kube.CustomResourceDefinition("helmreleases.helm.bitnami.com") + $.namespace {
    spec+: {
      group: "helm.bitnami.com",
      version: "v1",
      scope: "Namespaced",
      names: {
        kind: "HelmRelease",
        listKind: "HelmReleaseList",
        plural: "helmreleases",
        singular: "helmrelease",
      },
    },
  },

  tillerServiceAccount: kube.ServiceAccount("tiller") + $.namespace,
  tillerBinding: kube.ClusterRoleBinding("tiller-cluster-admin") {
    roleRef_: kube.ClusterRole("cluster-admin"),
    subjects_: [$.tillerServiceAccount],
  },
  local tillerDeploy = (import "tiller-deployment.jsonnet"),
  local tillerContainer = tillerDeploy.spec.template.spec.containers[0],
  tillerHelmCRD: tillerDeploy + $.namespace {
    spec+: {
      template+: {
        spec+: {
          serviceAccountName: $.tillerServiceAccount.metadata.name,
          containers: [tillerContainer {
            ports: [],  // Informational only, doesn't actually restrict access to :44134
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
  } + controllerOverlay,
}
