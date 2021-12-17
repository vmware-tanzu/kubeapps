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
import { IK8sList, IKubeState, IResource } from "./types";

export const APIBase = (cluster: string) => `api/clusters/${cluster}`;
export let WebSocketAPIBase: string;
if (window.location.protocol === "https:") {
  WebSocketAPIBase = `wss://${window.location.host}${window.location.pathname}`;
} else {
  WebSocketAPIBase = `ws://${window.location.host}${window.location.pathname}`;
}

// Kube is a lower-level class for interacting with the Kubernetes API. Use
// ResourceRef to interact with a single API resource rather than using Kube
// directly.
export class Kube {
  private static resourcesClient = () => new KubeappsGrpcClient().getResourcesServiceClientImpl();
  public static getResourceURL(
    cluster: string,
    apiVersion: string,
    resource: string,
    namespaced: boolean,
    namespace?: string,
    name?: string,
    query?: string,
  ) {
    let u = `${url.api.k8s.base(cluster)}/${
      apiVersion === "v1" || !apiVersion ? "api/v1" : `apis/${apiVersion}`
    }`;
    if (namespaced && namespace) {
      u += `/namespaces/${namespace}`;
    }
    u += `/${resource}`;
    if (name) {
      u += `/${name}`;
    }
    if (query) {
      u += `?${query}`;
    }
    return u;
  }

  public static async getResource(
    cluster: string,
    apiVersion: string,
    resource: string,
    namespaced: boolean,
    namespace?: string,
    name?: string,
    query?: string,
  ) {
    const { data } = await axiosWithAuth.get<IResource | IK8sList<IResource, {}>>(
      this.getResourceURL(cluster, apiVersion, resource, namespaced, namespace, name, query),
    );
    return data;
  }

  // getResources returns a subscription to an observable for resources from the server.
  public static getResources(
    pkgRef: InstalledPackageReference,
    refs: ResourceRef[],
    watch: boolean,
  ) {
    return this.resourcesClient().GetResources({
      installedPackageRef: pkgRef,
      resourceRefs: refs,
      watch,
    });
  }

  public static async getAPIGroups(cluster: string) {
    const { data: apiGroups } = await axiosWithAuth.get<any>(`${url.api.k8s.base(cluster)}/apis`);
    return apiGroups.groups;
  }

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
    const { data: coreResourceList } = await axiosWithAuth.get<any>(
      `${url.api.k8s.base(cluster)}/api/v1`,
    );
    coreResourceList.resources?.forEach((r: any) => addResource(r, "v1"));

    await Promise.all(
      groups.map(async (g: any) => {
        const groupVersion = g.preferredVersion.groupVersion;
        const { data: resourceList } = await axiosWithAuth.get<any>(
          `${url.api.k8s.base(cluster)}/apis/${groupVersion}`,
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
      const { data } = await axiosWithAuth.post<{ allowed: boolean }>(url.backend.canI(cluster), {
        group,
        resource,
        verb,
        namespace,
      });
      return data?.allowed ? data.allowed : false;
    } catch (e: any) {
      return false;
    }
  }

  public static async getServiceAccountNames(
    cluster: string,
    namespace: string,
  ): Promise<GetServiceAccountNamesResponse> {
    return await this.resourcesClient().GetServiceAccountNames({
      context: { cluster, namespace },
    } as GetServiceAccountNamesRequest);
  }
}
