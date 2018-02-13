import axios from "axios";

import * as urls from "../shared/url";
import { AppRepository } from "./AppRepository";
import { IAppRepository, IK8sList, IStatus } from "./types";

export class ServiceCatalog {
  public static async getCatalogRepo(): Promise<IAppRepository | undefined> {
    const repos = await AppRepository.list();
    const svcRepo = repos.items.find(repo => {
      return !!(
        repo.spec &&
        repo.spec.url &&
        repo.spec.url === "https://svc-catalog-charts.storage.googleapis.com"
      );
    });

    return svcRepo;
  }

  public static async installCatalog(name: string, namespace: string) {
    const url = "https://svc-catalog-charts.storage.googleapis.com";
    await AppRepository.create(name, url, namespace);
  }

  public static async getServiceClasses() {
    return ServiceCatalog.getItems<IServiceClass>("/clusterserviceclasses");
  }

  public static async getServiceBrokers() {
    return ServiceCatalog.getItems<IServiceBroker>("/clusterservicebrokers");
  }

  public static async getServicePlans() {
    return ServiceCatalog.getItems<IServicePlan>("/clusterserviceplans");
  }

  public static async getServiceBindings(): Promise<IServiceBinding[]> {
    const bindings = await ServiceCatalog.getItems<IServiceBinding>("/servicebindings");

    // initiate with undefined secrets
    for (const binding of bindings) {
      binding.spec = {
        ...binding.spec,
        secretDatabase: undefined,
        secretHost: undefined,
        secretPassword: undefined,
        secretPort: undefined,
        secretUsername: undefined,
      };
    }

    return Promise.all(
      bindings.map(binding => {
        const { secretName } = binding.spec;
        const { namespace } = binding.metadata;
        return axios
          .get<IK8sApiSecretResponse>(ServiceCatalog.secretEndpoint(namespace) + secretName)
          .then(response => {
            const { database, host, password, port, username } = response.data.data;
            const spec = {
              ...binding.spec,
              secretDatabase: atob(database),
              secretHost: atob(host),
              secretPassword: atob(password),
              secretPort: atob(port),
              secretUsername: atob(username),
            };
            return { ...binding, spec };
          })
          .catch(err => {
            // return with undefined secrets
            return { ...binding };
          });
      }),
    );
  }

  public static async provisionInstance(
    releaseName: string,
    namespace: string,
    className: string,
    planName: string,
    parameters: {},
  ) {
    const { data } = await axios.post<IStatus>(
      urls.api.serviceinstances.create(namespace),
      {
        apiVersion: "servicecatalog.k8s.io/v1beta1",
        kind: "ServiceInstance",
        metadata: {
          name: releaseName,
        },
        spec: {
          clusterServiceClassExternalName: className,
          clusterServicePlanExternalName: planName,
          parameters,
        },
      },
      {
        validateStatus: statusCode => true,
      },
    );

    if (data.status === "Failure") {
      throw new Error(data.message);
    }

    return data;
  }

  public static async deprovisionInstance(instance: IServiceInstance) {
    const { data } = await axios.delete("/api/kube" + instance.metadata.selfLink);

    if (data.status === "Failure") {
      throw new Error(data.message);
    }

    return data;
  }

  public static async syncBroker(broker: IServiceBroker) {
    const { data } = await axios.patch<IStatus>(
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

    if (data.status === "Failure") {
      throw new Error(data.message);
    }

    return data;
  }

  public static async getServiceInstances() {
    return ServiceCatalog.getItems<IServiceInstance>("/serviceinstances");
  }

  public static async isCatalogInstalled(): Promise<boolean> {
    try {
      const { status } = await axios.get(ServiceCatalog.endpoint);
      return status === 200;
    } catch (err) {
      return false;
    }
  }

  private static endpoint: string = "/api/kube/apis/servicecatalog.k8s.io/v1beta1";

  private static secretEndpoint = (namespace: string = "default") =>
    `/api/kube/api/v1/namespaces/${namespace}/secrets/`;

  private static async getItems<T>(endpoint: string): Promise<T[]> {
    const response = await axios.get<IK8sList<T, {}>>(ServiceCatalog.endpoint + endpoint);
    // const response = await axios.get<IK8sApiListResponse<T>>(ServiceCatalog.endpoint + endpoint);
    const json = response.data;
    return json.items;
  }
}

interface IK8sApiSecretResponse {
  kind: string;
  apiVersion: string;
  metadata: {
    selfLink: string;
    resourceVersion: string;
  };
  data: {
    database: string;
    host: string;
    password: string;
    port: string;
    username: string;
  };
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

export interface IServiceClass {
  metadata: {
    creationTimestamp: string;
    name: string;
    resourceVersion: string;
    selfLink: string;
    uid: string;
  };
  spec: {
    bindable: boolean;
    binding_retrievable: boolean;
    clusterServiceBrokerName: string;
    description: string;
    externalID: string;
    externalName: string;
    planUpdatable: boolean;
    tags: string[];
    externalMetadata?: {
      displayName: string;
      documentationUrl: string;
      imageUrl: string;
      longDescription: string;
      supportUrl: string;
    };
  };
  status: {
    removedFromBrokerCatalog: boolean;
  };
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
    free: boolean;
    clusterServiceClassRef: {
      name: string;
    };
  };
  status: {
    removedFromBrokerCatalog: boolean;
  };
}

export interface IServiceInstance {
  metadata: {
    name: string;
    namespace: string;
    selfLink: string;
    uid: string;
    resourceVersion: string;
    creationTimestamp: string;
    finalizers: string[];
    generation: number;
  };
  spec: {
    clusterServiceClassExternalName: string;
    clusterServicePlanExternalName: string;
    externalID: string;
    clusterServicePlanRef: {
      name: string;
    };
    clusterServiceClassRef: {
      name: string;
    };
  };
  status: { conditions: ICondition[] };
}

export interface IServiceBinding {
  metadata: {
    name: string;
    selfLink: string;
    uid: string;
    resourceVersion: string;
    creationTimestamp: string;
    finalizers: string[];
    generation: number;
    namespace: string;
  };
  spec: {
    externalID: string;
    instanceRef: {
      name: string;
    };
    secretName: string | undefined;
    secretDatabase: string | undefined;
    secretHost: string | undefined;
    secretPassword: string | undefined;
    secretPort: string | undefined;
    secretUsername: string | undefined;
  };
}
