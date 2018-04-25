local kube = import "kube.libsonnet";

local kubeless = import "vendor/kubeless/kubeless.jsonnet";

local controllerContainer = kubeless.controller.spec.template.spec.containers[0] + { image: "bitnami/kubeless-controller-manager:" + std.extVar("KUBELESS_VERSION") };
local config = kubeless.cfg + { data+: { "builder-image": "kubeless/function-image-builder:" + std.extVar("KUBELESS_VERSION") } };

kubeless {
  ns: kube.Namespace("kubeless"),
  controller+: {spec+: {template+: {spec+: { containers: [controllerContainer] } } } },
  cfg: config,
}
