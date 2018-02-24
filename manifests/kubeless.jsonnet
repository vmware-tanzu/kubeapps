local kube = import "kube.libsonnet";

local kubeless = import "vendor/kubeless/kubeless-rbac.jsonnet";

local controllerContainer = kubeless.controller.spec.template.spec.containers[0] + { image: "bitnami/kubeless-controller@sha256:939d64b5a50c36036738530d29af678752236e412d06c5eda1831be1a5b588e4" };

kubeless {
  ns: kube.Namespace("kubeless"),
  controller+: {spec+: {template+: {spec+: { containers: [controllerContainer] } } } },
}
