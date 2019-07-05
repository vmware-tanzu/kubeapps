import { axios } from "./AxiosInstance";
import { IChart } from "./types";

export default class Chart {
  public static async getReadme(id: string, version: string) {
    const url = `${Chart.APIEndpoint}/assets/${id}/versions/${version}/README.md`;
    const { data } = await axios.get<string>(url);
    return data;
  }

  public static async getValues(id: string, version: string) {
    const url = `${Chart.APIEndpoint}/assets/${id}/versions/${version}/values.yaml`;
    const { data } = await axios.get<string>(url);
    return data;
  }

  public static async exists(id: string, version: string, repo: string) {
    const url = `${Chart.APIEndpoint}/charts/${repo}/${id}/versions/${version}`;
    try {
      await axios.get<string>(url);
    } catch (e) {
      return false;
    }
    return true;
  }

  public static async listWithFilters(name: string, version: string, appVersion: string) {
    const url = `${
      Chart.APIEndpoint
    }/charts?name=${name}&version=${version}&appversion=${appVersion}`;
    const { data } = await axios.get<{ data: IChart[] }>(url);
    return data.data;
  }

  private static APIEndpoint: string = "api/chartsvc/v1";
}
