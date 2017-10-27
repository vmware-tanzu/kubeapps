local kube = import "kube.libsonnet";

{
  namespace:: {metadata+: {namespace: "kubeapps"}},

  ns: kube.Namespace($.namespace.metadata.namespace),

  // NB: these are left in their usual namespaces, to avoid forcing
  // non-default command line options onto client tools
  kubeless: (import "kubeless.jsonnet"),
  tiller: (import "tiller.jsonnet"),
  ssecrets: (import "sealed-secrets.jsonnet"),

  hub: (import "kubeapps-hub.jsonnet") + {
    namespace:: $.namespace,
    mongodb:: $.mongodb.svc,
  },

  mongodb: (import "mongodb.jsonnet") + {
    namespace:: $.namespace,
  },
}
