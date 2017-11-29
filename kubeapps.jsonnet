local kube = import "kube.libsonnet";
local kubecfg = import "kubecfg.libsonnet";

local host = null;
local tls = false;

{
  namespace:: {metadata+: {namespace: "kubeapps"}},

  ns: kube.Namespace($.namespace.metadata.namespace),

  // NB: these are left in their usual namespaces, to avoid forcing
  // non-default command line options onto client tools
  kubeless: (import "kubeless.jsonnet"),
  ssecrets: (import "sealed-secrets.jsonnet"),
  nginx: (import "ingress-nginx.jsonnet") {
    namespace:: $.namespace,
    controller+: {
      spec+: {
        template+: {
          spec+: {
            containers_+: {
              default+: {
                args_+: {
                  "ingress-class": "kubeapps-nginx",
                }
              }
            }
          }
        }
      }
    },
    service+: {
      spec+: {
        local maybe_https = if tls then [
          {name: "https", port: 443, protocol: "TCP"},
        ] else [],

        ports: [
          {name: "http", port: 80, protocol: "TCP"},
        ] + maybe_https,
      },
    },
  },

  kubelessui: (import "kubeless-ui.jsonnet") {
    namespace:: $.namespace,
  },

  dashboard: (import "kubeapps-dashboard.jsonnet") + {
    namespace:: $.namespace,
    mongodb_svc:: $.mongodb.svc,
    mongodb_secret:: $.mongodb.secret,
    ingress:: null,
    values+: {
      api+: {
        service+: {type: "ClusterIP"},
        // FIXME: api server downloads metadata/icons/etc for *every
        // chart* *before* it starts answering /healthz
        livenessProbe+: {initialDelaySeconds: 10*60},
      },
      ui+: {service+: {type: "ClusterIP"}},
      prerender+: {service+: {type: "ClusterIP"}},
    },

    // FIXME(gus): I think these are bugs in the monocular chart
    local readinessDelay(value) = {
      deploy+: {
        spec+: {
          template+: {
            spec+: {
              containers_+: {
                default+: {
                  readinessProbe+: {
                    initialDelaySeconds: value,
                  },
                },
              },
            },
          },
        },
      },
    },
    ui+: readinessDelay(0),
    api+: readinessDelay(0),
  },

  mongodb: (import "mongodb.jsonnet") {
    namespace:: $.namespace,
  },

  ingress: kube.Ingress("kubeapps") + $.namespace {
    metadata+: {
      annotations+: {
        "ingress.kubernetes.io/rewrite-target": "/",
        "kubernetes.io/ingress.class": "kubeapps-nginx",
        "ingress.kubernetes.io/ssl-redirect": std.toString(tls),
      },
    },
    spec+: {
      rules: [{
        http: {
          paths: [
            {path: "/", backend: $.dashboard.ui.svc.name_port},
            {path: "/api/", backend: $.dashboard.api.svc.name_port},
            {path: "/kubeless", backend: $.kubelessui.svc.name_port},
          ],
        },
        host: host,
      }],

      tls: if tls then [{
        secretName: $.ingressTls.metadata.name,
        hosts: host,
      }] else [],
    },
  },
}
