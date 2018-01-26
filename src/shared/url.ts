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
  charts: {
    base: "/api/chartsvc/v1",
    get: (id: string) => `${api.charts.base}/charts/${id}`,
    getReadme: (id: string, version: string) =>
      `${api.charts.base}/assets/${id}/versions/${version}/README.md`,
    list: (repo?: string) => `${api.charts.base}/charts${repo ? `/${repo}` : ""}`,
    listVersions: (id: string) => `${api.charts.get(id)}/versions`,
  },

  helmreleases: {
    create: (namespace = "default") =>
      `/api/kube/apis/helm.bitnami.com/v1/namespaces/${namespace}/helmreleases`,
  },
};
