import { AppRepository } from "./AppRepository";
import { axios } from "./Auth";
import { hapi } from "./hapi/release";
import { IAppOverview, IChartVersion } from "./types";

export const TILLER_PROXY_ROOT_URL = "api/tiller-deploy/v1";

export class App {
  public static getResourceURL(namespace?: string, name?: string, query?: string) {
    let url = TILLER_PROXY_ROOT_URL;
    if (namespace) {
      url += `/namespaces/${namespace}`;
    }
    url += "/releases";
    if (name) {
      url += `/${name}`;
    }
    if (query) {
      url += `?${query}`;
    }
    return url;
  }

  public static async create(
    releaseName: string,
    namespace: string,
    kubeappsNamespace: string,
    chartVersion: IChartVersion,
    values?: string,
  ) {
    const chartAttrs = chartVersion.relationships.chart.data;
    const repo = await AppRepository.get(chartAttrs.repo.name, kubeappsNamespace);
    const auth = repo.spec.auth;
    const endpoint = App.getResourceURL(namespace);
    const { data } = await axios.post(endpoint, {
      auth,
      chartName: chartAttrs.name,
      releaseName,
      repoUrl: chartAttrs.repo.url,
      values,
      version: chartVersion.attributes.version,
    });
    return data;
  }

  public static async upgrade(
    releaseName: string,
    namespace: string,
    kubeappsNamespace: string,
    chartVersion: IChartVersion,
    values?: string,
  ) {
    const chartAttrs = chartVersion.relationships.chart.data;
    const repo = await AppRepository.get(chartAttrs.repo.name, kubeappsNamespace);
    const auth = repo.spec.auth;
    const endpoint = App.getResourceURL(namespace, releaseName);
    const { data } = await axios.put(endpoint, {
      auth,
      chartName: chartAttrs.name,
      releaseName,
      repoUrl: chartAttrs.repo.url,
      values,
      version: chartVersion.attributes.version,
    });
    return data;
  }

  public static async delete(releaseName: string, namespace: string, purge: boolean) {
    let purgeQuery;
    if (purge) {
      purgeQuery = "purge=true";
    }
    const { data } = await axios.delete(App.getResourceURL(namespace, releaseName, purgeQuery));
    return data;
  }

  public static async listApps(namespace?: string, allStatuses?: boolean) {
    let query;
    if (allStatuses) {
      query = "statuses=all";
    }
    const { data } = await axios.get<{ data: IAppOverview[] }>(
      App.getResourceURL(namespace, undefined, query),
    );
    return data.data;
  }

  public static async getRelease(namespace: string, name: string) {
    const { data } = await axios.get<{ data: hapi.release.Release }>(
      this.getResourceURL(namespace, name),
    );
    return data.data;
  }
}
