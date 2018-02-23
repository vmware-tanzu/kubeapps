local kube = import "kube.libsonnet";
local kubecfg = import "kubecfg.libsonnet";

local labels = {
  app: "kubeapps-dashboard",
};

// ConfigMap, with a content-hash appended to the name.
local HashedConfigMap(name) = kube.ConfigMap(name) {
  local hash = std.substr(std.md5(std.toString($.data)), 0, 7),
  metadata+: {name: "%s-%s" % [super.name, hash]},
};

{
  namespace:: {metadata+: {namespace: "kubeapps"}},
  mongodb_svc:: error "a mongodb service is required",
  mongodb_secret:: error "a mongodb secret is required",

  local name = labels.app,
  local mongoDbHost = "%s.%s" % [$.mongodb_svc.metadata.name, $.mongodb_svc.metadata.namespace],

  tillerHelmCRD: (import "helm-crd.jsonnet") { namespace: $.namespace },

  ui: {
    svc: kube.Service(name + "-ui") + $.namespace {
      metadata+: {labels+: labels},
      target_pod: $.ui.deploy.spec.template,
    },

    deploy: kube.Deployment(name + "-ui") + $.namespace {
      spec+: {
        template+: {
          spec+: {
            containers_+: {
              default: kube.Container("dashboard") {
                image: "kubeapps/dashboard:v0.4.0",
                ports_: {
                  http: {containerPort: 8080, protocol: "TCP"},
                },
                livenessProbe: {
                  httpGet: {
                    path: "/",
                    port: 8080,
                  },
                  initialDelaySeconds: 60,
                  timeoutSeconds: 10,
                },
                readinessProbe: self.livenessProbe {
                  initialDelaySeconds: 0,
                  timeoutSeconds: 5,
                },
                volumeMounts_+: {
                  vhost: {mountPath: "/bitnami/nginx/conf/vhosts"},
                },
              },
            },
            volumes_+: {
              vhost: kube.ConfigMapVolume($.ui.vhost),
            },
          },
        },
      },
    },

    vhost: HashedConfigMap(name + "-ui-vhost") + $.namespace {
      metadata+: {labels+: labels},

      data+: {
        ui_port:: 8080,

        "vhost.conf": (importstr "kubeapps-ui-vhost.conf") % self,
      },
    },
  },

  apprepository: (import "apprepository.jsonnet"),
  chartsvc: (import "chartsvc.jsonnet") {
    mongodb_secret: $.mongodb_secret,
    mongodb_host: mongoDbHost,
  },
  kubeapi: (import "kube-api.jsonnet"),
}
