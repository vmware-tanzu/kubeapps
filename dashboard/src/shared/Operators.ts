import * as urls from "../shared/url";
import { axiosWithAuth } from "./AxiosInstance";

export class Operators {
  public static async isOLMInstalled() {
    try {
      const { status } = await axiosWithAuth.get(urls.api.operators.crd);
      return status === 200;
    } catch (err) {
      return false;
    }
  }
}
