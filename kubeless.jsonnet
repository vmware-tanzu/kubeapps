local kube = import "kube.libsonnet";

local kubeless = import "vendor/kubeless/kubeless-rbac.jsonnet";

local controllerContainer = kubeless.controller.spec.template.spec.containers[0] + { image: "bitnami/kubeless-controller@sha256:53592e0f023353665569313a1662a3aff18141e48caf4beca54d68436e71e0dc" };

kubeless {
  ns: kube.Namespace("kubeless"),
  controller+: {spec+: {template+: {spec+: { containers: [controllerContainer] } } } },
}
