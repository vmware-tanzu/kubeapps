local kube = import "kube.libsonnet";
local kubecfg = import "kubecfg.libsonnet";

{
  namespace:: {metadata+: {namespace: "kubeapps"}},

  applications: {
    // kubeapps-applications-read
    // Gives read-only access to Applications within a Namespace in Kubeapps.
    // Usage:
    //   Apply kubeapps-applications-read clusterrole to user/serviceaccount in the desired namespace
    read: kube.ClusterRole("kubeapps-applications-read") {
      rules: [
        {
          apiGroups: ["*"],
          resources: ["*"],
          verbs: ["list", "get", "watch"],
        },
      ],
    },
    // kubeapps-applications-write
    // Gives write access to Applications within a Namespace in Kubeapps.
    // Usage:
    //   Apply kubeapps-applications-write clusterrole to user/serviceaccount in the desired namespace.
    //   AND apply kubeapps-repositories-read role to user/serviceaccount in the kubeapps namespace.
    write: kube.ClusterRole("kubeapps-applications-write") {
      rules: [
        {
          apiGroups: ["*"],
          resources: ["*"],
          verbs: ["create", "patch", "update", "delete"],
        },
      ],
    },
  },

  functions: {
    // kubeapps-functions-read
    // Gives read-only access to Functions within a Namespace in Kubeapps.
    // Usage:
    //   Apply kubeapps-functions-read clusterrole to user/serviceaccount in the desired namespace
    //   AND apply kubeapps-kubeless-config-read to user/serviceaccount in the kubeless namespace.
    read: kube.ClusterRole("kubeapps-functions-read") {
      rules: [
        {
          apiGroups: ["kubeless.io"],
          resources: ["functions"],
          verbs: ["list", "get"],
        },
        {
          apiGroups: ["", "apps"],
          resources: ["pods", "deployments"],
          verbs: ["list", "watch"],
        },
        {
          apiGroups: [""],
          resources: ["pods/log"],
          verbs: ["get"],
        },
        // Allows Kubeapps to send GET/POST requests to function endpoints.
        {
          apiGroups: [""],
          resources: ["services/proxy"],
          verbs: ["get", "create"],
        },
      ],
    },
    readKubelessConfig: kube.Role("kubeapps-kubeless-config-read") {
      metadata+: {
        namespace: "kubeless",
      },
      rules: [
        {
          apiGroups: [""],
          resources: ["configmaps"],
          resourceNames: ["kubeless-config"],
          verbs: ["get"],
        },
      ],
    },
    // kubeapps-functions-write
    // Gives write access to Functions within a Namespace in Kubeapps.
    // Usage:
    //   Apply kubeapps-functions-write clusterrole to user/serviceaccount in the desired namespace.
    write: kube.ClusterRole("kubeapps-functions-write") {
      rules: [
        {
          apiGroups: ["kubeless.io"],
          resources: ["functions"],
          verbs: ["create", "update", "delete"],
        },
      ],
    },
  },

  serviceCatalog: {
    // kubeapps-service-catalog-read
    // Gives read-only access to Service Instances and Bindings within a Namespace in Kubeapps.
    // Usage:
    //   Apply kubeapps-service-catalog-read clusterrole to user/serviceaccount in the desired namespace
    //   AND apply kubeapps-service-catalog-browse to user/serviceaccount in all namespaces.
    browse: kube.ClusterRole("kubeapps-service-catalog-browse") {
      rules: [
        {
          apiGroups: ["servicecatalog.k8s.io"],
          resources: ["clusterservicebrokers", "clusterserviceclasses", "clusterserviceplans"],
          verbs: ["list"],
        },
      ],
    },
    read: kube.ClusterRole("kubeapps-service-catalog-read") {
      rules: [
        {
          apiGroups: ["servicecatalog.k8s.io"],
          resources: ["serviceinstances", "servicebindings"],
          verbs: ["list", "get"],
        },
        // Allows viewing Service Binding credentials.
        {
          apiGroups: [""],
          resources: ["secrets"],
          verbs: ["get"],
        },
      ],
    },
    // kubeapps-service-catalog-write
    // Gives write access to Service Instances and Bindings within a Namespace in Kubeapps.
    // Usage:
    //   Apply kubeapps-service-catalog-write clusterrole to user/serviceaccount in the desired namespace.
    write: kube.ClusterRole("kubeapps-service-catalog-write") {
      rules: [
        {
          apiGroups: ["servicecatalog.k8s.io"],
          resources: ["serviceinstances", "servicebindings"],
          verbs: ["create", "delete"],
        },
      ],
    },
    // kubeapps-service-catalog-admin
    // Gives admin access for the Service Broker configuration page in Kubeapps.
    // Usage:
    //   Apply kubeapps-service-catalog-admin clusterrole to user/serviceaccount in all namespaces.
    admin: kube.ClusterRole("kubeapps-service-catalog-admin") {
      rules: [
        {
          apiGroups: ["servicecatalog.k8s.io"],
          resources: ["clusterservicebrokers"],
          verbs: ["patch"],
        },
      ],
    },
  },

  repositories: {
    // kubeapps-repositories-read
    // Gives read-only access to App Repositories in Kubeapps.
    // Usage:
    //   Apply kubeapps-repositories-read role to user/serviceaccount in the kubeapps namespace.
    read: kube.Role("kubeapps-repositories-read") + $.namespace {
      rules: [
        {
          apiGroups: ["kubeapps.com"],
          resources: ["apprepositories"],
          verbs: ["list", "get"],
        }
      ],
    },
    // kubeapps-repositories-write
    // Gives write access to App Repositories in Kubeapps.
    // Usage:
    //   Apply kubeapps-repositories-write role to user/serviceaccount in the kubeapps namespace.
    write: kube.Role("kubeapps-repositories-write") + $.namespace {
      rules: [
        {
          apiGroups: ["kubeapps.com"],
          resources: ["apprepositories"],
          verbs: ["get", "create", "update", "delete"],
        },
        {
          apiGroups: [""],
          resources: ["secrets"],
          verbs: ["create"],
        },
      ],
    },
  },
}
