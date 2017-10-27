local kube = import "kube.libsonnet";

// Should probably copy/inline this, but for now just assume that
// kubeless is checked out somewhere nearby
local kubeless = import "../../kubeless/kubeless/kubeless-rbac.jsonnet";

kubeless {
  ns: kube.Namespace("kubeless"),
}
