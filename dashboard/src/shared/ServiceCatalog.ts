import { JSONSchema6 } from "json-schema";
import * as url from "shared/url";
import { axiosWithAuth } from "./AxiosInstance";
import { IClusterServiceClass } from "./ClusterServiceClass";
import { APIBase } from "./Kube";
import { IServiceInstance } from "./ServiceInstance";
import { IK8sList, IStatus } from "./types";

export class ServiceCatalog {
  public static async getServiceClasses(cluster: string) {
    return this.getItems<IClusterServiceClass>(cluster, "clusterserviceclasses");
  }

  public static async getServiceBrokers(cluster: string) {
    return this.getItems<IServiceBroker>(cluster, "clusterservicebrokers");
  }

  public static async getServicePlans(cluster: string) {
    return this.getItems<IServicePlan>(cluster, "clusterserviceplans");
  }

  public static async deprovisionInstance(cluster: string, instance: IServiceInstance) {
    const { data } = await axiosWithAuth.delete(`${APIBase(cluster)}${instance.metadata.selfLink}`);
    return data;
  }

  public static async syncBroker(cluster: string, broker: IServiceBroker) {
    const { data } = await axiosWithAuth.patch<IStatus>(
      url.api.k8s.clusterservicebrokers.sync(cluster, broker),
      {
        spec: {
          relistRequests: broker.spec.relistRequests + 1,
        },
      },
      {
        headers: { "Content-Type": "application/merge-patch+json" },
        validateStatus: () => true,
      },
    );
    return data;
  }

  public static async isCatalogInstalled(cluster: string): Promise<boolean> {
    try {
      const { status } = await axiosWithAuth.get(this.endpoint(cluster));
      return status === 200;
    } catch (e: any) {
      return false;
    }
  }

  public static async getItems<T>(
    cluster: string,
    resource: string,
    namespace?: string,
  ): Promise<T[]> {
    const response = await axiosWithAuth.get<IK8sList<T, {}>>(
      this.endpoint(cluster) + (namespace ? `/namespaces/${namespace}` : "") + `/${resource}`,
    );
    const json = response.data;
    return json.items;
  }

  private static endpoint(cluster: string): string {
    return `${APIBase(cluster)}/apis/servicecatalog.k8s.io/v1beta1`;
  }
}
export interface IK8sApiListResponse<T> {
  kind: string;
  apiVersion: string;
  metadata: {
    selfLink: string;
    resourceVersion: string;
  };
  items: T[];
}

export interface ICondition {
  type: string;
  status: string;
  lastTransitionTime: string;
  reason: string;
  message: string;
}

export interface IServiceBroker {
  metadata: {
    name: string;
    selfLink: string;
    uid: string;
    resourceVersion: string;
    generation: number;
    creationTimestamp: string;
    finalizers: string[];
  };
  spec: {
    url: string;
    authInfo: any; // Look into
    relistBehavior: string;
    relistDuration: string;
    relistRequests: number;
  };
  status: {
    conditions: ICondition[];
    reconciledGeneration: number;
    lastCatalogRetrievalTime: string;
  };
}

export interface IServicePlan {
  metadata: {
    name: string;
    selfLink: string;
    uid: string;
    resourceVersion: string;
    creationTimestamp: string;
  };
  spec: {
    clusterServiceBrokerName: string;
    externalName: string;
    externalID: string;
    description: string;
    externalMetadata?: {
      displayName: string;
      bullets: string[];
    };
    instanceCreateParameterSchema?: JSONSchema6;
    serviceBindingCreateParameterSchema?: JSONSchema6;
    free: boolean;
    clusterServiceClassRef: {
      name: string;
    };
  };
  status: {
    removedFromBrokerCatalog: boolean;
  };
}
