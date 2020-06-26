import { JSONSchema4 } from "json-schema";
import { axiosWithAuth } from "./AxiosInstance";
import { IChart, IChartVersion } from "./types";
import * as URL from "./url";

export default class Chart {
  public static async fetchCharts(namespace: string, repo: string) {
    const { data } = await axiosWithAuth.get<{ data: IChart[] }>(
      URL.api.charts.list(namespace, repo),
    );
    return data.data;
  }

  public static async fetchChartVersions(namespace: string, id: string): Promise<IChartVersion[]> {
    const { data } = await axiosWithAuth.get<{ data: IChartVersion[] }>(
      URL.api.charts.listVersions(namespace, id),
    );
    return data.data;
  }

  public static async getChartVersion(namespace: string, id: string, version: string) {
    const { data } = await axiosWithAuth.get<{ data: IChartVersion }>(
      URL.api.charts.getVersion(namespace, id, version),
    );
    return data.data;
  }

  public static async getReadme(namespace: string, id: string, version: string) {
    const { data } = await axiosWithAuth.get<string>(
      URL.api.charts.getReadme(namespace, id, version),
    );
    return data;
  }

  public static async getValues(namespace: string, id: string, version: string) {
    const { data } = await axiosWithAuth.get<string>(
      URL.api.charts.getValues(namespace, id, version),
    );
    return data;
  }

  public static async getSchema(namespace: string, id: string, version: string) {
    const { data } = await axiosWithAuth.get<JSONSchema4>(
      URL.api.charts.getSchema(namespace, id, version),
    );
    return data;
  }

  public static async listWithFilters(
    namespace: string,
    name: string,
    version: string,
    appVersion: string,
  ) {
    const url = `${
      Chart.APIEndpoint
    }/ns/${namespace}/charts?name=${name}&version=${encodeURIComponent(
      version,
    )}&appversion=${appVersion}`;
    const { data } = await axiosWithAuth.get<{ data: IChart[] }>(url);
    return data.data;
  }

  private static APIEndpoint: string = "api/assetsvc/v1";
}
