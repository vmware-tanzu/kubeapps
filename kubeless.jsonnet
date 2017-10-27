local kube = import "kube.libsonnet";

local kubeless = import "vendor/kubeless/kubeless-rbac.jsonnet";

kubeless {
  ns: kube.Namespace("kubeless"),
}
