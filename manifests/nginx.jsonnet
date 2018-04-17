local kube = import "kube.libsonnet";
local kubecfg = import "kubecfg.libsonnet";

local labels = {
  app: "kubeapps",
};

// ConfigMap, with a content-hash appended to the name.
local HashedConfigMap(name) = kube.ConfigMap(name) {
  local hash = std.substr(std.md5(std.toString($.data)), 0, 7),
  metadata+: {name: "%s-%s" % [super.name, hash]},
};

{
  namespace:: {metadata+: {namespace: "kubeapps"}},

  local name = labels.app,

  vhost: HashedConfigMap(name + "-vhost") + $.namespace {
    metadata+: {labels+: labels},

    data+: {
      ui_port:: 8080,

      "vhost.conf": (importstr "nginx-vhost.conf") % self,
    },
  },

  deploy: kube.Deployment(name) + $.namespace {
    metadata+: {labels+: labels},
    spec+: {
      template+: {
        spec+: {
          containers_+: {
            default: kube.Container("nginx") {
              image: "bitnami/nginx:1.12",
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
            vhost: kube.ConfigMapVolume($.vhost),
          },
        },
      },
    },
  },

  service: kube.Service(name) + $.namespace {
    metadata+: {labels+: labels},
    target_pod: $.deploy.spec.template,
  },
}
