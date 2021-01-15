import { JSONSchema4 } from "json-schema";
import { axiosWithAuth } from "./AxiosInstance";
import { IChart, IChartCategory, IChartVersion } from "./types";
import * as URL from "./url";

export default class Chart {
  public static async fetchCharts(cluster: string, namespace: string, repo: string) {
    const { data } = await axiosWithAuth.get<{ data: IChart[] }>(
      URL.api.charts.list(cluster, namespace, repo),
    );
    return data.data;
  }

  public static async fetchChartCategories(cluster: string, namespace: string, repos?: string) {
    const { data } = await axiosWithAuth.get<{ data: IChartCategory[] }>(
      URL.api.charts.getChartCategories(cluster, namespace, repos),
    );
    return data.data;
  }

  public static async fetchChartVersions(
    cluster: string,
    namespace: string,
    id: string,
  ): Promise<IChartVersion[]> {
    const { data } = await axiosWithAuth.get<{ data: IChartVersion[] }>(
      URL.api.charts.listVersions(cluster, namespace, id),
    );
    return data.data;
  }

  public static async getChartVersion(
    cluster: string,
    namespace: string,
    id: string,
    version: string,
  ) {
    const { data } = await axiosWithAuth.get<{ data: IChartVersion }>(
      URL.api.charts.getVersion(cluster, namespace, id, version),
    );
    return data.data;
  }

  public static async getReadme(cluster: string, namespace: string, id: string, version: string) {
    const { data } = await axiosWithAuth.get<string>(
      URL.api.charts.getReadme(cluster, namespace, id, version),
    );
    return data;
  }

  public static async getValues(cluster: string, namespace: string, id: string, version: string) {
    const { data } = await axiosWithAuth.get<string>(
      URL.api.charts.getValues(cluster, namespace, id, version),
    );
    return data;
  }

  public static async getSchema(cluster: string, namespace: string, id: string, version: string) {
    const { data } = await axiosWithAuth.get<JSONSchema4>(
      URL.api.charts.getSchema(cluster, namespace, id, version),
    );
    return data;
  }

  public static async listWithFilters(
    cluster: string,
    namespace: string,
    name: string,
    version: string,
    appVersion: string,
  ) {
    const url = `${URL.api.charts.base(cluster, namespace)}/charts?name=${encodeURIComponent(
      name,
    )}&version=${encodeURIComponent(version)}&appversion=${appVersion}`;
    const { data } = await axiosWithAuth.get<{ data: IChart[] }>(url);
    return data.data;
  }
}
