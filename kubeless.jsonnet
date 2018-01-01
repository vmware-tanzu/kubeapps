local kube = import "kube.libsonnet";

local kubeless = import "vendor/kubeless/kubeless-rbac.jsonnet";

local controllerContainer = kubeless.controller.spec.template.spec.containers[0] + { image: "bitnami/kubeless-controller@sha256:9dadf6d1707655c3eaf724e23d6c9cef046dbe377e6f8f7341d5e2ef4f3c0dde" };

kubeless {
  ns: kube.Namespace("kubeless"),
  controller+: {spec+: {template+: {spec+: { containers: [controllerContainer] } } } },
}
