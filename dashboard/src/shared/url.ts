import { IServiceBroker } from "./ServiceCatalog";
import { IChartVersion, IRepo } from "./types";

export const app = {
  apps: {
    new: (
      cluster: string,
      namespace: string,
      cv: IChartVersion,
      version: string,
      globalNamespace: string,
    ) => {
      const repoNamespace = cv.relationships.chart.data.repo.namespace;
      const newSegment = globalNamespace !== repoNamespace ? "new" : "new-from-global";
      return `/c/${cluster}/ns/${namespace}/apps/${newSegment}/${
        cv.relationships.chart.data.repo.name
      }/${encodeURIComponent(cv.relationships.chart.data.name)}/versions/${version}`;
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
    get: (
      cluster: string,
      namespace: string,
      chartName: string,
      repo: IRepo,
      globalNamespace: string,
    ) => {
      const chartsSegment = globalNamespace === repo.namespace ? "global-charts" : "charts";
      return `/c/${cluster}/ns/${namespace}/${chartsSegment}/${repo.name}/${encodeURIComponent(
        chartName,
      )}`;
    },
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
  return namespace ? `namespaces/${namespace}/` : "";
}

export const backend = {
  namespaces: {
    list: (cluster: string) => `api/v1/clusters/${cluster}/namespaces`,
  },
  apprepositories: {
    base: (cluster: string, namespace: string) =>
      `api/v1/clusters/${cluster}/${withNS(namespace)}apprepositories`,
    create: (cluster: string, namespace: string) =>
      backend.apprepositories.base(cluster, namespace),
    list: (cluster: string, namespace: string) => backend.apprepositories.base(cluster, namespace),
    validate: (cluster: string, namespace: string) =>
      `${backend.apprepositories.base(cluster, namespace)}/validate`,
    delete: (cluster: string, namespace: string, name: string) =>
      `${backend.apprepositories.base(cluster, namespace)}/${name}`,
    refresh: (cluster: string, namespace: string, name: string) =>
      `${backend.apprepositories.base(cluster, namespace)}/${name}/refresh`,
    update: (cluster: string, namespace: string, name: string) =>
      `${backend.apprepositories.base(cluster, namespace)}/${name}`,
  },
  canI: (cluster: string) => `api/v1/clusters/${cluster}/can-i`,
};

export const kubeops = {
  releases: {
    list: (cluster: string, namespace: string) =>
      `api/kubeops/v1/clusters/${cluster}/namespaces/${namespace}/releases`,
    listAll: (cluster: string) => `api/kubeops/v1/clusters/${cluster}/releases`,
    get: (cluster: string, namespace: string, name: string) =>
      `${kubeops.releases.list(cluster, namespace)}/${name}`,
  },
};

export const api = {
  charts: {
    base: (cluster: string, namespace: string) =>
      `api/assetsvc/v1/clusters/${cluster}/namespaces/${namespace}`,
    get: (cluster: string, namespace: string, id: string) =>
      `${api.charts.base(cluster, namespace)}/charts/${id}`,
    getVersion: (cluster: string, namespace: string, id: string, version: string) =>
      `${api.charts.get(cluster, namespace, id)}/versions/${encodeURIComponent(version)}`,
    list: (
      cluster: string,
      namespace: string,
      repos: string,
      page: number,
      size: number,
      query?: string,
    ) =>
      `${api.charts.base(cluster, namespace)}/charts?page=${page}&size=${size}${
        query ? "&q=" + query : ""
      }${repos ? `&repos=${repos}` : ""}`,
    getChartCategories: (cluster: string, namespace: string) =>
      `${api.charts.base(cluster, namespace)}/charts/categories`,
    listVersions: (cluster: string, namespace: string, id: string) =>
      `${api.charts.get(cluster, namespace, id)}/versions`,
    getReadme: (cluster: string, namespace: string, id: string, version: string) =>
      `${api.charts.base(cluster, namespace)}/assets/${id}/versions/${encodeURIComponent(
        version,
      )}/README.md`,
    getValues: (cluster: string, namespace: string, id: string, version: string) =>
      `${api.charts.base(cluster, namespace)}/assets/${id}/versions/${encodeURIComponent(
        version,
      )}/values.yaml`,
    getSchema: (cluster: string, namespace: string, id: string, version: string) =>
      `${api.charts.base(cluster, namespace)}/assets/${id}/versions/${encodeURIComponent(
        version,
      )}/values.schema.json`,
  },

  // URLs which are accessing the k8s API server directly are grouped together
  // so we can clearly differentiate and possibly begin to remove.
  // Note that this list is not yet exhaustive (search for APIBase to find other call-sites which
  // access the k8s api server directly).
  k8s: {
    base: (cluster: string) => `api/clusters/${cluster}`,
    namespaces: (cluster: string) => `${api.k8s.base(cluster)}/api/v1/namespaces`,
    namespace: (cluster: string, namespace: string) =>
      namespace ? `${api.k8s.namespaces(cluster)}/${namespace}` : `${api.k8s.base(cluster)}/api/v1`,
    // clusterservicebrokers and operators operate on the default cluster only, currently.
    clusterservicebrokers: {
      sync: (cluster: string, broker: IServiceBroker) =>
        `${api.k8s.base(cluster)}/apis/servicecatalog.k8s.io/v1beta1/clusterservicebrokers/${
          broker.metadata.name
        }`,
    },
    operators: {
      operators: (cluster: string, namespace: string) =>
        `${api.k8s.base(cluster)}/apis/packages.operators.coreos.com/v1/${withNS(
          namespace,
        )}packagemanifests`,
      operator: (cluster: string, namespace: string, name: string) =>
        `${api.k8s.base(
          cluster,
        )}/apis/packages.operators.coreos.com/v1/namespaces/${namespace}/packagemanifests/${name}`,
      clusterServiceVersions: (cluster: string, namespace: string) =>
        `${api.k8s.base(cluster)}/apis/operators.coreos.com/v1alpha1/${withNS(
          namespace,
        )}clusterserviceversions`,
      clusterServiceVersion: (cluster: string, namespace: string, name: string) =>
        `${api.k8s.base(
          cluster,
        )}/apis/operators.coreos.com/v1alpha1/namespaces/${namespace}/clusterserviceversions/${name}`,
      resources: (cluster: string, namespace: string, apiVersion: string, resource: string) =>
        `${api.k8s.base(cluster)}/apis/${apiVersion}/${withNS(namespace)}${resource}`,
      resource: (
        cluster: string,
        namespace: string,
        apiVersion: string,
        resource: string,
        name: string,
      ) =>
        `${api.k8s.base(cluster)}/apis/${apiVersion}/namespaces/${namespace}/${resource}/${name}`,
      operatorGroups: (cluster: string, namespace: string) =>
        `${api.k8s.base(
          cluster,
        )}/apis/operators.coreos.com/v1/namespaces/${namespace}/operatorgroups`,
      subscriptions: (cluster: string, namespace: string) =>
        `${api.k8s.base(
          cluster,
        )}/apis/operators.coreos.com/v1alpha1/namespaces/${namespace}/subscriptions`,
      subscription: (cluster: string, namespace: string, name: string) =>
        `${api.k8s.base(
          cluster,
        )}/apis/operators.coreos.com/v1alpha1/namespaces/${namespace}/subscriptions/${name}`,
    },
    secrets: (cluster: string, namespace: string, fieldSelector?: string) =>
      `${api.k8s.namespace(cluster, namespace)}/secrets${
        fieldSelector ? `?fieldSelector=${encodeURIComponent(fieldSelector)}` : ""
      }`,
    secret: (cluster: string, namespace: string, name: string) =>
      `${api.k8s.secrets(cluster, namespace)}/${name}`,
  },

  operators: {
    operatorIcon: (cluster: string, namespace: string, name: string) =>
      `api/v1/clusters/${cluster}/namespaces/${namespace}/operator/${name}/logo`,
  },

  kubeappsapis: "/apis",
};
