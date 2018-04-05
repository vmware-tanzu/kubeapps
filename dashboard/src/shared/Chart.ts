import axios from "axios";

// import { IFunction, IFunctionList, IResource, IStatus } from "./types";

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

  private static APIEndpoint: string = "/api/chartsvc/v1";
}
