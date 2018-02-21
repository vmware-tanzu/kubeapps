local kube = import "kube.libsonnet";
local kubecfg = import "kubecfg.libsonnet";

local host = null;
local tls = false;
local kubeless = import "kubeless.jsonnet";
local ssecrets = import "sealed-secrets.jsonnet";

local labels = {
  metadata+: {
    labels+: {
      "created-by": "kubeapps"
    }
  }
};
// Some manifests are nested deeper than the root (e.g. dashboard.api.deploy)
// so we need to make sure we're only applying the labels to objects that have
// the manifest key
local labelify(src) = if std.objectHas(src, "metadata") then src + labels else src;
local labelifyEach(src) = {
  [k]: labelify(src[k]) for k in std.objectFields(src)
};

{
  namespace:: {metadata+: {namespace: "kubeapps"}},

  ns: kube.Namespace($.namespace.metadata.namespace) + labels,

  // NB: these are left in their usual namespaces, to avoid forcing
  // non-default command line options onto client tools
  kubeless: labelifyEach(kubeless),
  ssecrets: [s + labels for s in ssecrets],
  nginx_:: (import "ingress-nginx.jsonnet") {
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
  nginx: labelifyEach($.nginx_),

  kubelessui_:: (import "kubeless-ui.jsonnet") {
    namespace:: $.namespace,
  },
  kubelessui: labelifyEach($.kubelessui_),

  dashboard_:: (import "kubeapps-dashboard.jsonnet") {
    namespace:: $.namespace,
    mongodb_svc:: $.mongodb_.svc,
    mongodb_secret:: $.mongodb_.secret,
    ingress:: null,
    values+:: {
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
  },
  dashboard: labelifyEach($.dashboard_) {
    ui: labelifyEach($.dashboard_.ui),
    apprepository: labelifyEach($.dashboard_.apprepository) {
      apprepos: labelifyEach($.dashboard_.apprepository.apprepos),
    },
    chartsvc: labelifyEach($.dashboard_.chartsvc),
    kubeapi: labelifyEach($.dashboard_.kubeapi),
    tillerHelmCRD: labelifyEach($.dashboard_.tillerHelmCRD),
  },

  mongodb_:: (import "mongodb.jsonnet") {
    namespace:: $.namespace,
  },
  mongodb: labelifyEach($.mongodb_),

  ingress: kube.Ingress("kubeapps") + $.namespace + labels {
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
            {path: "/api/chartsvc", backend: $.dashboard.chartsvc.service.name_port},
            {path: "/api/kube", backend: $.dashboard.kubeapi.service.name_port},
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
