import placeholder from "../placeholder.png";
import * as urls from "../shared/url";
import { axiosWithAuth } from "./AxiosInstance";
import {
  IClusterServiceVersion,
  IK8sList,
  IPackageManifest,
  IPackageManifestChannel,
  IResource,
} from "./types";

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

  public static async createOperator(
    namespace: string,
    name: string,
    channel: string,
    installPlanApproval: string,
    csv: string,
  ) {
    // First create the OperatorGroup if needed
    await this.createOperatorGroupIfNotExists(namespace);
    // Now create the subscription
    const { data: result } = await axiosWithAuth.post<IResource>(
      urls.api.operators.subscription(namespace, name),
      {
        apiVersion: "operators.coreos.com/v1alpha1",
        kind: "Subscription",
        metadata: {
          name,
          namespace,
        },
        spec: {
          channel,
          installPlanApproval,
          name,
          source: "operatorhubio-catalog",
          sourceNamespace: "olm",
          startingCSV: csv,
        },
      },
    );
    return result;
  }

  public static getDefaultChannel(operator: IPackageManifest) {
    return operator.status.channels.find(ch => ch.name === operator.status.defaultChannel);
  }

  public static global(channel?: IPackageManifestChannel) {
    return !!channel?.currentCSVDesc.installModes.find(m => m.type === "AllNamespaces")?.supported;
  }

  private static async createOperatorGroupIfNotExists(namespace: string) {
    if (namespace === "operators") {
      // The opertors ns already have an operatorgroup
      return;
    }
    const { data } = await axiosWithAuth.get<IK8sList<IResource, {}>>(
      urls.api.operators.operatorGroups(namespace),
    );
    if (data.items.length > 0) {
      // An operatorgroup already exists, do nothing
      return;
    }
    const { data: result } = await axiosWithAuth.post<IK8sList<IResource, {}>>(
      urls.api.operators.operatorGroups(namespace),
      {
        apiVersion: "operators.coreos.com/v1",
        kind: "OperatorGroup",
        metadata: {
          generateName: "default-",
          namespace,
        },
        spec: {
          targetNamespaces: [namespace],
        },
      },
    );
    return result;
  }
}

export function getIcon(csv: IClusterServiceVersion) {
  if (csv.spec.icon && csv.spec.icon.length > 0) {
    return `data:${csv.spec.icon[0].mediatype};base64,${csv.spec.icon[0].base64data}`;
  }
  return placeholder;
}

export function findOwnedKind(csv: IClusterServiceVersion, kind: string) {
  if (
    csv.spec.customresourcedefinitions?.owned &&
    csv.spec.customresourcedefinitions?.owned.length > 0
  ) {
    return csv.spec.customresourcedefinitions.owned.find(c => c.kind === kind);
  }
  return;
}
