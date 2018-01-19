export const api = {
  charts: {
    get: (id: string) => `/api/chartsvc/v1/charts/${id}`,
    list: (repo?: string) => `/api/chartsvc/v1/charts${repo ? `/${repo}` : ""}`,
  },

  helmreleases: {
    create: (namespace = "default") =>
      `/api/kube/apis/helm.bitnami.com/v1/namespaces/${namespace}/helmreleases`,
  },
};
