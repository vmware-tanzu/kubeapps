import { JSONSchema6 } from "json-schema";
import * as urls from "../shared/url";
import { axiosWithAuth } from "./AxiosInstance";
import { IClusterServiceClass } from "./ClusterServiceClass";
import { APIBase } from "./Kube";
import { IServiceInstance } from "./ServiceInstance";
import { IK8sList, IStatus } from "./types";

export class ServiceCatalog {
  public static async getServiceClasses() {
    return this.getItems<IClusterServiceClass>("clusterserviceclasses");
  }

  public static async getServiceBrokers() {
    return this.getItems<IServiceBroker>("clusterservicebrokers");
  }

  public static async getServicePlans() {
    return this.getItems<IServicePlan>("clusterserviceplans");
  }

  public static async deprovisionInstance(instance: IServiceInstance) {
    const { data } = await axiosWithAuth.delete(`${APIBase}${instance.metadata.selfLink}`);
    return data;
  }

  public static async syncBroker(broker: IServiceBroker) {
    const { data } = await axiosWithAuth.patch<IStatus>(
      urls.api.clusterservicebrokers.sync(broker),
      {
        spec: {
          relistRequests: broker.spec.relistRequests + 1,
        },
      },
      {
        headers: { "Content-Type": "application/merge-patch+json" },
        validateStatus: statusCode => true,
      },
    );
    return data;
  }

  public static async isCatalogInstalled(): Promise<boolean> {
    try {
      const { status } = await axiosWithAuth.get(this.endpoint);
      return status === 200;
    } catch (err) {
      return false;
    }
  }

  public static async getItems<T>(resource: string, namespace?: string): Promise<T[]> {
    const response = await axiosWithAuth.get<IK8sList<T, {}>>(
      this.endpoint + (namespace ? `/namespaces/${namespace}` : "") + `/${resource}`,
    );
    const json = response.data;
    return json.items;
  }

  private static endpoint: string = `${APIBase}/apis/servicecatalog.k8s.io/v1beta1`;
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
