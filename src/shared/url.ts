export const api = {
  charts: {
    list: (repo?: string) => `/api/chartsvc/v1/charts${repo ? `/${repo}` : ''}`,
    get: (id: string) => `/api/chartsvc/v1/charts/${id}`
  },

  helmreleases: {
    create: (namespace = 'default') => `/api/kube/apis/helm.bitnami.com/v1/namespaces/${namespace}/helmreleases`,
  },
};
