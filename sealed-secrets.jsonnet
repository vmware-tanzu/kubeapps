local controller = import "../../bitnami/sealed-secrets/controller.jsonnet";

// For <k8s 1.7, we want to use TPRs:
//local crd = import "../../bitnami/sealed-secrets/sealedsecret-tpr.jsonnet";
local crd = import "../../bitnami/sealed-secrets/sealedsecret-crd.jsonnet";

controller + crd
