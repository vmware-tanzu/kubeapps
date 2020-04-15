import * as urls from "../shared/url";
import { axiosWithAuth } from "./AxiosInstance";
import { IClusterServiceVersion, IK8sList, IPackageManifest, IResource } from "./types";

export class Operators {
  public static async isOLMInstalled(namespace: string) {
    const { status } = await axiosWithAuth.get(urls.api.operators.operators(namespace));
    return status === 200;
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
    // Global operators are installed in the "operators" namespace
    const reqNamespace = namespace === "_all" ? "operators" : namespace;
    const { data } = await axiosWithAuth.get<IK8sList<IClusterServiceVersion, {}>>(
      urls.api.operators.clusterServiceVersions(reqNamespace),
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

  public static async listResources(namespace: string, apiVersion: string, resource: string) {
    const { data } = await axiosWithAuth.get<IK8sList<IResource, {}>>(
      urls.api.operators.resources(namespace, apiVersion, resource),
    );
    return data;
  }

  public static async getResource(
    namespace: string,
    apiVersion: string,
    crd: string,
    name: string,
  ) {
    const { data } = await axiosWithAuth.get<IResource>(
      urls.api.operators.resource(namespace, apiVersion, crd, name),
    );
    return data;
  }

  public static async deleteResource(
    namespace: string,
    apiVersion: string,
    plural: string,
    name: string,
  ) {
    const { data } = await axiosWithAuth.delete<any>(
      urls.api.operators.resource(namespace, apiVersion, plural, name),
    );
    return data;
  }

  public static async updateResource(
    namespace: string,
    apiVersion: string,
    resource: string,
    name: string,
    body: object,
  ) {
    const { data } = await axiosWithAuth.put<IResource>(
      urls.api.operators.resource(namespace, apiVersion, resource, name),
      body,
    );
    return data;
  }
}
