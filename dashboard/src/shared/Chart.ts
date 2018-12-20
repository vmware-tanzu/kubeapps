import axios from "axios";

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

  private static APIEndpoint: string = "api/chartsvc/v1";
}
