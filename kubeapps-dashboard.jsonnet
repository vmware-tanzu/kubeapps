// Roughly transcribed from
// https://github.com/kubeapps/hub/blob/master/chart/kubeapps-hub/values.yaml
// as of df6f032ac98f8a604407529601947c81c22bc776

local kube = import "kube.libsonnet";
local kubecfg = import "kubecfg.libsonnet";

local labels = {
  app: "kubeapps-dashboard",
};

local valuesDefault = kubecfg.parseYaml(importstr "kubeapps-dashboard-values.yaml")[0];

// ConfigMap, with a content-hash appended to the name.
local HashedConfigMap(name) = kube.ConfigMap(name) {
  local hash = std.substr(std.md5(std.toString($.data)), 0, 7),
  metadata+: {name: "%s-%s" % [super.name, hash]},
};

local serviceDeployFromValues(parentName, componentName, values) = {
  local name = "%s-%s" % [parentName, componentName],

  svc: kube.Service(name) {
    metadata+: {labels+: labels},
    target_pod: $.deploy.spec.template,
    spec+: {
      type: values.service.type,
      ports: [{
        port: values.service.externalPort,
        targetPort: values.service.internalPort,
        protocol: "TCP",
        name: values.service.name,
      }],
    },
  },

  deploy: kube.Deployment(name) {
    metadata+: {labels+: labels},
    spec+: {
      replicas: values.replicaCount,
      template+: {
        spec+: {
          default_container: "default",
          containers_+: {
            default: kube.Container(componentName) {
              image: "%s:%s" % [values.image.repository, values.image.tag],
              imagePullPolicy: values.image.pullPolicy,
              ports: [{containerPort: values.service.internalPort}],
              resources: values.resources,
            },
          },
        },
      },
    },
  },
};

{
  namespace:: {metadata+: {namespace: "kubeapps"}},
  mongodb_svc:: error "a mongodb service is required",
  mongodb_secret:: error "a mongodb secret is required",
  values:: valuesDefault,

  local name = labels.app,
  local mongoDbHost = "%s.%s" % [$.mongodb_svc.metadata.name, $.mongodb_svc.metadata.namespace],

  ingress: kube.Ingress(name) + $.namespace {
    metadata+: {
      labels+: labels + $.values.ingress.labels,
      annotations+: $.values.ingress.annotations,
    },

    spec+: {
      rules: [{
        http: {
          paths: [
            {path: "/", backend: $.ui.svc.name_port},
            {path: "/api/", backend: $.api.svc.name_port},
          ],
        },
        host: host,
      } for host in $.values.ingress.hosts],

      tls: if std.objectHas($.values.ingress, "tls") then [{
        secretName: $.values.ingress.tls.secretName,
        hosts: $.values.ingress.hosts,
      }] else [],
    },
  },

  tillerHelmCRD: (import "helm-crd.jsonnet") { namespace: $.namespace },

  ui: serviceDeployFromValues(name, "ui", $.values.ui) {
    config: HashedConfigMap(name + "-ui-config") + $.namespace {
      metadata+: {labels+: labels},

      data: {
        window_monocular:: {
          overrides: {
            googleAnalyticsId: $.values.ui.googleAnalyticsId,
            appName: $.values.ui.appName,
            backendHostname: $.values.ui.backendHostname,
            releasesEnabled: $.values.api.config.releasesEnabled,
          },
        },
        "overrides.js": "window.monocular = " + kubecfg.manifestJson(self.window_monocular)
      },
    },

    vhost: HashedConfigMap(name + "-ui-vhost") + $.namespace {
      metadata+: {labels+: labels},

      data+: {
        prerender_svc:: $.prerender.svc.metadata.name,
        ui_port:: $.values.ui.service.internalPort,

        "vhost.conf": (importstr "kubeapps-ui-vhost.conf") % self,
      },
    },

    svc+: $.namespace,

    deploy+: $.namespace {
      spec+: {
        template+: {
          spec+: {
            containers_+: {
              default+: {
                livenessProbe: {
                  httpGet: {
                    path: "/",
                    port: $.values.ui.service.internalPort,
                  },
                  initialDelaySeconds: 60,
                  timeoutSeconds: 10,
                },
                readinessProbe: self.livenessProbe {
                  initialDelaySeconds: 30,
                  timeoutSeconds: 5,
                },
                volumeMounts_+: {
                  vhost: {mountPath: "/bitnami/nginx/conf/vhosts"},
                  config: {mountPath: "/app/assets/js"},
                },
              },
            },
            volumes_+: {
              vhost: kube.ConfigMapVolume($.ui.vhost),
              config: kube.ConfigMapVolume($.ui.config),
            },
          },
        },
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
