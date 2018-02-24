local kube = import "kube.libsonnet";

local labels = {app: "mongodb"};

{
  namespace:: {metadata+: {namespace: "mongodb"}},

  secret:: kube.Secret("mongodb") + $.namespace {
    metadata+: {labels+: labels},
    data_+: {
      "mongodb-root-password": error "Value provided elsewhere",
    },
  },

  svc: kube.Service("mongodb") + $.namespace {
    target_pod: $.mongodb.spec.template,
  },

  pvc: kube.PersistentVolumeClaim("mongodb-data") + $.namespace {
    storage: "8Gi",
  },

  mongodb: kube.Deployment("mongodb") + $.namespace {
    metadata+: {labels+: labels},

    spec+: {
      template+: {
        spec+: {
          containers_+: {
            default: kube.Container("mongodb") {
              image: "bitnami/mongodb:3.4.9-r1",
              env_+: {
                MONGODB_ROOT_PASSWORD: kube.SecretKeyRef($.secret, "mongodb-root-password"),
              },
              ports_+: {
                mongodb: {containerPort: 27017},
              },
              livenessProbe: {
                exec: {
                  command: ["mongo", "--eval", "db.adminCommand('ping')"],
                },
                initialDelaySeconds: 30,
                timeoutSeconds: 5,
              },
              readinessProbe: self.livenessProbe {
                initialDelaySeconds: 5,
                timeoutSeconds: 1,
              },
              volumeMounts_+: {
                data: {mountPath: "/bitnami/mongodb"},
              },
              resources: {
                requests: {memory: "256Mi", cpu: "100m"},
              },
            },
          },
          volumes_+: {
            data: kube.PersistentVolumeClaimVolume($.pvc),
          },
        },
      },
    },
  },
}
