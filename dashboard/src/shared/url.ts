import { IServiceBroker } from "./ServiceCatalog";
import { IChartVersion } from "./types";

export const app = {
  charts: {
    version: (cv: IChartVersion) =>
      `/charts/${cv.relationships.chart.data.repo.name}/${
        cv.relationships.chart.data.name
      }/versions/${cv.attributes.version}`,
  },
};

export const api = {
  apprepostories: {
    base: `/api/kube/apis/kubeapps.com/v1alpha1`,
    create: (namespace = "default") =>
      `${api.apprepostories.base}/namespaces/${namespace}/apprepositories`,
  },

  charts: {
    base: "/api/chartsvc/v1",
    get: (id: string) => `${api.charts.base}/charts/${id}`,
    getReadme: (id: string, version: string) =>
      `${api.charts.base}/assets/${id}/versions/${version}/README.md`,
    getValues: (id: string, version: string) =>
      `${api.charts.base}/assets/${id}/versions/${version}/values.yaml`,
    getVersion: (id: string, version: string) =>
      `${api.charts.base}/charts/${id}/versions/${version}`,
    list: (repo?: string) => `${api.charts.base}/charts${repo ? `/${repo}` : ""}`,
    listVersions: (id: string) => `${api.charts.get(id)}/versions`,
  },

  // /api/kube exposes kubectl add ?watch=true
  helmreleases: {
    create: (namespace = "default") =>
      `/api/kube/apis/helm.bitnami.com/v1/namespaces/${namespace}/helmreleases`,
    list: (namespace?: string) =>
      `/api/kube/apis/helm.bitnami.com/v1/${
        namespace ? `namespaces/${namespace}/` : ""
      }/helmreleases`,
    listDetails: (releaseNames: string[]) =>
      `/api/kube/api/v1/namespaces/kubeapps/configmaps?labelSelector=NAME in (${releaseNames.join(
        ",",
      )})`,
  },
  serviceinstances: {
    base: `/api/kube/apis/servicecatalog.k8s.io/v1beta1`,
    create: (namespace = "default") =>
      `${api.serviceinstances.base}/namespaces/${namespace}/serviceinstances`,
  },

  clusterservicebrokers: {
    base: `/api/kube/apis/servicecatalog.k8s.io/v1beta1`,
    sync: (broker: IServiceBroker) =>
      `${api.clusterservicebrokers.base}/clusterservicebrokers/${broker.metadata.name}`,
  },
};
