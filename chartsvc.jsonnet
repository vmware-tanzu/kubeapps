local kube = import "kube.libsonnet";
local kubecfg = import "kubecfg.libsonnet";

local labels = {
  app: "chartsvc",
};

{
  namespace:: {metadata+: {namespace: "kubeapps"}},
  mongodb_secret:: error "a mongodb secret is required",
  mongodb_host:: error "a mongodb host is required",

  deployment: kube.Deployment("chartsvc") + $.namespace {
    metadata+: {labels+: labels},
    spec+: {
      template+: {
        spec+: {
          containers_+: {
            default: kube.Container("chartsvc") {
              image: "kubeapps/chartsvc:v0.2.0",
              env_+: {
                MONGO_PASSWORD: kube.SecretKeyRef($.mongodb_secret, "mongodb-root-password"),
              },
              command: ["/chartsvc"],
              args_: {
                "mongo-user": "root",
                "mongo-url": $.mongodb_host,
              },
              ports_: {
                http: {containerPort: 8080}
              },
            },
          },
        },
      },
    },
  },

  service: kube.Service("chartsvc") + $.namespace {
    target_pod: $.deployment.spec.template,
  },
}
