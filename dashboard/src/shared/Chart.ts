import { JSONSchema4 } from "json-schema";
import { axiosWithAuth } from "./AxiosInstance";
import { IChart, IChartVersion } from "./types";
import * as URL from "./url";

export default class Chart {
  public static async fetchCharts(repo: string) {
    const { data } = await axiosWithAuth.get<{ data: IChart[] }>(URL.api.charts.list(repo));
    return data.data;
  }

  public static async fetchChartVersions(id: string) {
    const { data } = await axiosWithAuth.get<{ data: IChartVersion[] }>(
      URL.api.charts.listVersions(id),
    );
    return data.data;
  }

  public static async getChartVersion(id: string, version: string) {
    const { data } = await axiosWithAuth.get<{ data: IChartVersion }>(
      URL.api.charts.getVersion(id, version),
    );
    return data.data;
  }

  public static async getReadme(id: string, version: string) {
    const { data } = await axiosWithAuth.get<string>(URL.api.charts.getReadme(id, version));
    return data;
  }

  public static async getValues(id: string, version: string) {
    const { data } = await axiosWithAuth.get<string>(URL.api.charts.getValues(id, version));
    return data;
  }

  public static async getSchema(id: string, version: string) {
    const { data } = await axiosWithAuth.get<JSONSchema4>(URL.api.charts.getSchema(id, version));
    return data;
  }

  public static async exists(id: string, version: string, repo: string) {
    try {
      await axiosWithAuth.get(URL.api.charts.getVersion(`${repo}/${id}`, version));
    } catch (e) {
      return false;
    }
    return true;
  }

  public static async listWithFilters(name: string, version: string, appVersion: string) {
    const url = `${Chart.APIEndpoint}/charts?name=${name}&version=${encodeURIComponent(
      version,
    )}&appversion=${appVersion}`;
    const { data } = await axiosWithAuth.get<{ data: IChart[] }>(url);
    return data.data;
  }

  private static APIEndpoint: string = "api/chartsvc/v1";
}
