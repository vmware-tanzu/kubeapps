local kubecfg = import "kubecfg.libsonnet";

// sealed-secrets/controller.jsonnet isn't designed to be used outside
// the sealed-secrets build, so we just use the built YAML files here :(

local controller = kubecfg.parseYaml(importstr "sealedsecret-controller.yaml");

// For <k8s 1.7, we want to use TPRs:
//local crd = kubecfg.parseYaml(importstr "sealedsecret-tpr.yaml");
local crd = kubecfg.parseYaml(importstr "sealedsecret-crd.yaml");

// NB: this expression is an array, because YAML
controller + crd
