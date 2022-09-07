// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import {
  InstalledPackageReference,
  ResourceRef,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import {
  GetServiceAccountNamesRequest,
  GetServiceAccountNamesResponse,
} from "gen/kubeappsapis/plugins/resources/v1alpha1/resources";
import * as url from "shared/url";
import { axiosWithAuth } from "./AxiosInstance";
import { KubeappsGrpcClient } from "./KubeappsGrpcClient";
import { IKubeState } from "./types";
import { convertGrpcAuthError } from "./utils";

// Kube is a lower-level class for interacting with the Kubernetes API. Use
// ResourceRef to interact with a single API resource rather than using Kube
// directly.
export class Kube {
  public static resourcesServiceClient = () =>
    new KubeappsGrpcClient().getResourcesServiceClientImpl();

  // getResources returns a subscription to an observable for resources from the server.
  public static getResources(
    pkgRef: InstalledPackageReference,
    refs: ResourceRef[],
    watch: boolean,
  ) {
    return this.resourcesServiceClient().GetResources({
      installedPackageRef: pkgRef,
      resourceRefs: refs,
      watch,
    });
  }

  // TODO(agamez): Migrate API call, see #4785
  public static async getAPIGroups(cluster: string) {
    const { data: apiGroups } = await axiosWithAuth.get<any>(url.api.k8s.apis(cluster));
    return apiGroups.groups;
  }

  // TODO(agamez): Migrate API call, see #4785
  public static async getResourceKinds(cluster: string, groups: any[]) {
    const result: IKubeState["kinds"] = {};
    const addResource = (r: any, version: string) => {
      // Exclude subresources
      if (!r.name.includes("/")) {
        result[r.kind] = {
          apiVersion: version,
          plural: r.name,
          namespaced: r.namespaced,
        };
      }
    };
    // Handle v1 separately
    const { data: coreResourceList } = await axiosWithAuth.get<any>(url.api.k8s.v1(cluster));
    coreResourceList.resources?.forEach((r: any) => addResource(r, "v1"));

    await Promise.all(
      groups.map(async (g: any) => {
        const groupVersion = g.preferredVersion.groupVersion;
        const { data: resourceList } = await axiosWithAuth.get<any>(
          url.api.k8s.groupVersion(cluster, groupVersion),
        );
        resourceList.resources?.forEach((r: any) => addResource(r, groupVersion));
      }),
    );
    return result;
  }

  public static async canI(
    cluster: string,
    group: string,
    resource: string,
    verb: string,
    namespace: string,
  ) {
    try {
      if (!cluster) {
        return false;
      }
      // TODO(rcastelblanq) Migrate the CanI endpoint to a proper RBAC/Auth plugin
      const response = await this.resourcesServiceClient().CanI({
        context: { cluster, namespace },
        group: group,
        resource: resource,
        verb: verb,
      });
      return response ? response.allowed : false;
    } catch (e: any) {
      return false;
    }
  }

  public static async getServiceAccountNames(
    cluster: string,
    namespace: string,
  ): Promise<GetServiceAccountNamesResponse> {
    return await this.resourcesServiceClient()
      .GetServiceAccountNames({
        context: { cluster, namespace },
      } as GetServiceAccountNamesRequest)
      .catch((e: any) => {
        throw convertGrpcAuthError(e);
      });
  }
}
