import { IServiceBroker } from "./ServiceCatalog";
import { IChartVersion, IRepo } from "./types";

export const app = {
  apps: {
    new: (cluster: string, namespace: string, cv: IChartVersion, version: string) => {
      const repoNamespace = cv.relationships.chart.data.repo.namespace;
      const newSegment = repoNamespace === namespace ? "new" : "new-from-global";
      return `/c/${cluster}/ns/${namespace}/apps/${newSegment}/${cv.relationships.chart.data.repo.name}/${cv.relationships.chart.data.name}/versions/${version}`;
    },
    list: (cluster: string, namespace: string) => `/c/${cluster}/ns/${namespace}/apps`,
    get: (cluster: string, namespace: string, releaseName: string) =>
      `${app.apps.list(cluster, namespace)}/${releaseName}`,
    upgrade: (cluster: string, namespace: string, releaseName: string) =>
      `${app.apps.get(cluster, namespace, releaseName)}/upgrade`,
  },
  catalog: (cluster: string, namespace: string) => `/c/${cluster}/ns/${namespace}/catalog`,
  repo: (cluster: string, namespace: string, repo: string) =>
    `${app.catalog(cluster, namespace)}/${repo}`,
  servicesInstances: (namespace: string) => `/ns/${namespace}/services/instances`,
  charts: {
    get: (cluster: string, namespace: string, chartName: string, repo: IRepo) => {
      const chartsSegment = namespace !== repo?.namespace ? "global-charts" : "charts";
      return `/c/${cluster}/ns/${namespace}/${chartsSegment}/${repo.name}/${chartName}`;
    },
    version: (
      cluster: string,
      namespace: string,
      chartName: string,
      chartVersion: string,
      repo: IRepo,
    ) => `${app.charts.get(cluster, namespace, chartName, repo)}/versions/${chartVersion}`,
  },
  operators: {
    view: (cluster: string, namespace: string, name: string) =>
      `/c/${cluster}/ns/${namespace}/operators/${name}`,
    list: (cluster: string, namespace: string) => `/c/${cluster}/ns/${namespace}/operators`,
    new: (cluster: string, namespace: string, name: string) =>
      `/c/${cluster}/ns/${namespace}/operators/new/${name}`,
  },
  operatorInstances: {
    view: (
      cluster: string,
      namespace: string,
      csvName: string,
      crdName: string,
      resourceName: string,
    ) => `/c/${cluster}/ns/${namespace}/operators-instances/${csvName}/${crdName}/${resourceName}`,
    update: (
      cluster: string,
      namespace: string,
      csvName: string,
      crdName: string,
      instanceName: string,
    ) =>
      `/c/${cluster}/ns/${namespace}/operators-instances/${csvName}/${crdName}/${instanceName}/update`,
    new: (cluster: string, namespace: string, csvName: string, crdName: string) =>
      `/c/${cluster}/ns/${namespace}/operators-instances/new/${csvName}/${crdName}`,
  },
  config: {
    apprepositories: (cluster: string, namespace: string) =>
      `/c/${cluster}/ns/${namespace}/config/repos`,
    brokers: (cluster: string) => `/c/${cluster}/config/brokers`,
    operators: (cluster: string, namespace: string) => `/c/${cluster}/ns/${namespace}/operators`,
  },
};

function withNS(namespace: string) {
  return namespace === "_all" ? "" : `namespaces/${namespace}/`;
}

export const backend = {
  namespaces: {
    list: (cluster: string) => `api/v1/clusters/${cluster}/namespaces`,
  },
  apprepositories: {
    base: (namespace: string) => `api/v1/namespaces/${namespace}/apprepositories`,
    create: (namespace: string) => backend.apprepositories.base(namespace),
    validate: () => `${backend.apprepositories.base("kubeapps")}/validate`,
    delete: (name: string, namespace: string) =>
      `${backend.apprepositories.base(namespace)}/${name}`,
    update: (namespace: string, name: string) =>
      `${backend.apprepositories.base(namespace)}/${name}`,
  },
};

export const kubeops = {
  releases: {
    list: (cluster: string, namespace: string) =>
      `api/tiller-deploy/v1/clusters/${cluster}/namespaces/${namespace}/releases`,
    listAll: (cluster: string) => `api/tiller-deploy/v1/clusters/${cluster}/releases`,
    get: (cluster: string, namespace: string, name: string) =>
      `${kubeops.releases.list(cluster, namespace)}/${name}`,
  },
};

export const api = {
  charts: {
    base: "api/assetsvc/v1",
    get: (namespace: string, id: string) => `${api.charts.base}/ns/${namespace}/charts/${id}`,
    getReadme: (namespace: string, id: string, version: string) =>
      `${api.charts.base}/ns/${namespace}/assets/${id}/versions/${encodeURIComponent(
        version,
      )}/README.md`,
    getValues: (namespace: string, id: string, version: string) =>
      `${api.charts.base}/ns/${namespace}/assets/${id}/versions/${encodeURIComponent(
        version,
      )}/values.yaml`,
    getSchema: (namespace: string, id: string, version: string) =>
      `${api.charts.base}/ns/${namespace}/assets/${id}/versions/${encodeURIComponent(
        version,
      )}/values.schema.json`,
    getVersion: (namespace: string, id: string, version: string) =>
      `${api.charts.base}/ns/${namespace}/charts/${id}/versions/${encodeURIComponent(version)}`,
    list: (namespace: string, repo?: string) =>
      `${api.charts.base}/ns/${namespace}/charts${repo ? `/${repo}` : ""}`,
    listVersions: (namespace: string, id: string) => `${api.charts.get(namespace, id)}/versions`,
  },

  // URLs which are accessing the k8s API server directly are grouped together
  // so we can clearly differentiate and possibly begin to remove.
  // Note that this list is not yet exhaustive (search for APIBase to find other call-sites which
  // access the k8s api server directly).
  k8s: {
    base: (cluster: string) => `api/clusters/${cluster}`,
    namespaces: (cluster: string) => `${api.k8s.base(cluster)}/api/v1/namespaces`,
    namespace: (cluster: string, namespace: string) =>
      `${api.k8s.namespaces(cluster)}/${namespace}`,
    // clusterservicebrokers and operators operate on the default cluster only, currently.
    clusterservicebrokers: {
      sync: (broker: IServiceBroker) =>
        `${api.k8s.base("default")}/apis/servicecatalog.k8s.io/v1beta1/clusterservicebrokers/${
          broker.metadata.name
        }`,
    },
    operators: {
      operators: (namespace: string) =>
        `${api.k8s.base("default")}/apis/packages.operators.coreos.com/v1/${withNS(
          namespace,
        )}packagemanifests`,
      operator: (namespace: string, name: string) =>
        `${api.k8s.base(
          "default",
        )}/apis/packages.operators.coreos.com/v1/namespaces/${namespace}/packagemanifests/${name}`,
      clusterServiceVersions: (namespace: string) =>
        `${api.k8s.base("default")}/apis/operators.coreos.com/v1alpha1/${withNS(
          namespace,
        )}clusterserviceversions`,
      clusterServiceVersion: (namespace: string, name: string) =>
        `${api.k8s.base(
          "default",
        )}/apis/operators.coreos.com/v1alpha1/namespaces/${namespace}/clusterserviceversions/${name}`,
      resources: (namespace: string, apiVersion: string, resource: string) =>
        `${api.k8s.base("default")}/apis/${apiVersion}/${withNS(namespace)}${resource}`,
      resource: (namespace: string, apiVersion: string, resource: string, name: string) =>
        `${api.k8s.base("default")}/apis/${apiVersion}/namespaces/${namespace}/${resource}/${name}`,
      operatorGroups: (namespace: string) =>
        `${api.k8s.base(
          "default",
        )}/apis/operators.coreos.com/v1/namespaces/${namespace}/operatorgroups`,
      subscription: (namespace: string, name: string) =>
        `${api.k8s.base(
          "default",
        )}/apis/operators.coreos.com/v1alpha1/namespaces/${namespace}/subscriptions/${name}`,
    },
    secrets: (cluster: string, namespace: string) =>
      `${api.k8s.namespace(cluster, namespace)}/secrets`,
    secret: (cluster: string, namespace: string, name: string) =>
      `${api.k8s.secrets(cluster, namespace)}/${name}`,
  },

  operators: {
    operatorIcon: (namespace: string, name: string) =>
      `api/v1/namespaces/${namespace}/operator/${name}/logo`,
  },
};
