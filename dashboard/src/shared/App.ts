import * as url from "shared/url";
import { axiosWithAuth } from "./AxiosInstance";
import { hapi } from "./hapi/release";
import { IAppOverview, IChartVersion } from "./types";

export const KUBEOPS_ROOT_URL = "api/kubeops/v1";

export class App {
  public static async create(
    cluster: string,
    namespace: string,
    releaseName: string,
    chartNamespace: string,
    chartVersion: IChartVersion,
    values?: string,
  ) {
    const chartAttrs = chartVersion.relationships.chart.data;
    const endpoint = url.kubeops.releases.list(cluster, namespace);
    const { data } = await axiosWithAuth.post(endpoint, {
      appRepositoryResourceName: chartAttrs.repo.name,
      appRepositoryResourceNamespace: chartNamespace,
      chartName: decodeURIComponent(chartAttrs.name),
      releaseName,
      values,
      version: chartVersion.attributes.version,
    });
    return data;
  }

  public static async upgrade(
    cluster: string,
    namespace: string,
    releaseName: string,
    chartNamespace: string,
    chartVersion: IChartVersion,
    values?: string,
  ) {
    const chartAttrs = chartVersion.relationships.chart.data;
    const endpoint = url.kubeops.releases.get(cluster, namespace, releaseName);
    const { data } = await axiosWithAuth.put(endpoint, {
      appRepositoryResourceName: chartAttrs.repo.name,
      appRepositoryResourceNamespace: chartNamespace,
      chartName: decodeURIComponent(chartAttrs.name),
      releaseName,
      values,
      version: chartVersion.attributes.version,
    });
    return data;
  }

  public static async rollback(
    cluster: string,
    namespace: string,
    releaseName: string,
    revision: number,
  ) {
    const endpoint = url.kubeops.releases.get(cluster, namespace, releaseName);
    const { data } = await axiosWithAuth.put(
      endpoint,
      {},
      {
        params: {
          action: "rollback",
          revision,
        },
      },
    );
    return data;
  }

  public static async delete(
    cluster: string,
    namespace: string,
    releaseName: string,
    purge: boolean,
  ) {
    let endpoint = url.kubeops.releases.get(cluster, namespace, releaseName);
    if (purge) {
      endpoint += "?purge=true";
    }
    const { data } = await axiosWithAuth.delete(endpoint);
    return data;
  }

  public static async listApps(cluster: string, namespace?: string) {
    let endpoint = namespace
      ? url.kubeops.releases.list(cluster, namespace)
      : url.kubeops.releases.listAll(cluster);
    endpoint += "?statuses=all";
    const { data } = await axiosWithAuth.get<{ data: IAppOverview[] }>(endpoint);
    return data.data;
  }

  public static async getRelease(cluster: string, namespace: string, name: string) {
    const { data } = await axiosWithAuth.get<{ data: hapi.release.Release }>(
      url.kubeops.releases.get(cluster, namespace, name),
    );
    return data.data;
  }
}
