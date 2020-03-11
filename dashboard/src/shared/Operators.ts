import * as urls from "../shared/url";
import { axiosWithAuth } from "./AxiosInstance";
import { IK8sList, IPackageManifest } from "./types";

export class Operators {
  public static async isOLMInstalled() {
    try {
      const { status } = await axiosWithAuth.get(urls.api.operators.crd);
      return status === 200;
    } catch (err) {
      return false;
    }
  }

  public static async getOperators(namespace: string) {
    const { data } = await axiosWithAuth.get<IK8sList<IPackageManifest, {}>>(
      urls.api.operators.operators(namespace),
    );
    return data.items;
  }

  public static async getOperator(namespace: string, name: string) {
    const { data } = await axiosWithAuth.get<IPackageManifest>(
      urls.api.operators.operator(namespace, name),
    );
    return data;
  }
}
