import * as urls from "../shared/url";
import { axiosWithAuth } from "./AxiosInstance";
import { IClusterServiceVersion, IK8sList, IPackageManifest, IResource } from "./types";

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

  public static async getCSVs(namespace: string) {
    const { data } = await axiosWithAuth.get<IK8sList<IClusterServiceVersion, {}>>(
      urls.api.operators.clusterServiceVersions(namespace),
    );
    return data.items;
  }

  public static async getCSV(namespace: string, name: string) {
    const { data } = await axiosWithAuth.get<IClusterServiceVersion>(
      urls.api.operators.clusterServiceVersion(namespace, name),
    );
    return data;
  }

  public static async createResource(
    namespace: string,
    apiVersion: string,
    resource: string,
    body: object,
  ) {
    const { data } = await axiosWithAuth.post<IResource>(
      urls.api.operators.resources(namespace, apiVersion, resource),
      body,
    );
    return data;
  }
}
