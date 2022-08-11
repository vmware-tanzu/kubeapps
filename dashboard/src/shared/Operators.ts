// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import * as url from "shared/url";
import placeholder from "icons/placeholder.svg";
import { axiosWithAuth } from "./AxiosInstance";
import {
  IClusterServiceVersion,
  IK8sList,
  IPackageManifest,
  IPackageManifestChannel,
  IResource,
} from "./types";

export class Operators {
  public static async isOLMInstalled(cluster: string, namespace: string) {
    const { status } = await axiosWithAuth.get<any>(
      url.api.k8s.operators.operators(cluster, namespace),
    );
    return status === 200;
  }

  public static async getOperators(cluster: string, namespace: string) {
    const { data } = await axiosWithAuth.get<IK8sList<IPackageManifest, {}>>(
      url.api.k8s.operators.operators(cluster, namespace),
    );
    return data.items;
  }

  public static async getOperator(cluster: string, namespace: string, name: string) {
    const { data } = await axiosWithAuth.get<IPackageManifest>(
      url.api.k8s.operators.operator(cluster, namespace, name),
    );
    return data;
  }

  public static async getCSVs(cluster: string, namespace: string) {
    // Global operators are installed in the "operators" namespace
    const { data } = await axiosWithAuth.get<IK8sList<IClusterServiceVersion, {}>>(
      url.api.k8s.operators.clusterServiceVersions(cluster, namespace),
    );
    return data.items;
  }

  public static async getCSV(cluster: string, namespace: string, name: string) {
    const { data } = await axiosWithAuth.get<IClusterServiceVersion>(
      url.api.k8s.operators.clusterServiceVersion(cluster, namespace, name),
    );
    return data;
  }

  public static async createResource(
    cluster: string,
    namespace: string,
    apiVersion: string,
    resource: string,
    body: object,
  ) {
    const { data } = await axiosWithAuth.post<IResource>(
      url.api.k8s.operators.resources(cluster, namespace, apiVersion, resource),
      body,
    );
    return data;
  }

  public static async listResources(
    cluster: string,
    namespace: string,
    apiVersion: string,
    resource: string,
  ) {
    const { data } = await axiosWithAuth.get<IK8sList<IResource, {}>>(
      url.api.k8s.operators.resources(cluster, namespace, apiVersion, resource),
    );
    return data;
  }

  public static async getResource(
    cluster: string,
    namespace: string,
    apiVersion: string,
    crd: string,
    name: string,
  ) {
    const { data } = await axiosWithAuth.get<IResource>(
      url.api.k8s.operators.resource(cluster, namespace, apiVersion, crd, name),
    );
    return data;
  }

  public static async deleteResource(
    cluster: string,
    namespace: string,
    apiVersion: string,
    plural: string,
    name: string,
  ) {
    const { data } = await axiosWithAuth.delete<any>(
      url.api.k8s.operators.resource(cluster, namespace, apiVersion, plural, name),
    );
    return data;
  }

  public static async updateResource(
    cluster: string,
    namespace: string,
    apiVersion: string,
    resource: string,
    name: string,
    body: object,
  ) {
    const { data } = await axiosWithAuth.put<IResource>(
      url.api.k8s.operators.resource(cluster, namespace, apiVersion, resource, name),
      body,
    );
    return data;
  }

  public static async createOperator(
    cluster: string,
    namespace: string,
    name: string,
    channel: string,
    installPlanApproval: string,
    csv: string,
  ) {
    // First create the OperatorGroup if needed
    await this.createOperatorGroupIfNotExists(cluster, namespace);
    // Now create the subscription
    const { data: result } = await axiosWithAuth.post<IResource>(
      url.api.k8s.operators.subscription(cluster, namespace, name),
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

  public static async listSubscriptions(cluster: string, namespace: string) {
    const { data } = await axiosWithAuth.get<IK8sList<IResource, {}>>(
      url.api.k8s.operators.subscriptions(cluster, namespace),
    );
    return data;
  }

  public static getDefaultChannel(operator: IPackageManifest) {
    return operator.status.channels.find(ch => ch.name === operator.status.defaultChannel);
  }

  public static global(channel?: IPackageManifestChannel) {
    return !!channel?.currentCSVDesc.installModes.find(m => m.type === "AllNamespaces")?.supported;
  }

  private static async createOperatorGroupIfNotExists(cluster: string, namespace: string) {
    if (namespace === "operators") {
      // The opertors ns already have an operatorgroup
      return;
    }
    const { data } = await axiosWithAuth.get<IK8sList<IResource, {}>>(
      url.api.k8s.operators.operatorGroups(cluster, namespace),
    );
    if (data.items.length > 0) {
      // An operatorgroup already exists, do nothing
      return;
    }
    const { data: result } = await axiosWithAuth.post<IK8sList<IResource, {}>>(
      url.api.k8s.operators.operatorGroups(cluster, namespace),
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
