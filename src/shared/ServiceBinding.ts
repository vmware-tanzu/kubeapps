import axios from "axios";
import { ICondition, ServiceCatalog } from "./ServiceCatalog";

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
  status: {
    conditions: ICondition[];
    asyncOpInProgress: boolean;
    currentOperation: string;
    reconciledGeneration: number;
    operationStartTime: string;
    externalProperties: {};
    orphanMitigationInProgress: boolean;
    unbindStatus: string;
  };
}

export class ServiceBinding {
  public static async create(
    bindingName: string,
    instanceRefName: string,
    namespace: string = "default",
  ) {
    const url = ServiceBinding.getLink(namespace);
    const { data } = await axios.post<IServiceBinding>(url, {
      metadata: {
        name: bindingName,
      },
      spec: {
        instanceRef: {
          name: instanceRefName,
        },
      },
    });
    return data;
  }

  public static async delete(name: string, namespace: string = "default") {
    const url = this.getLink(namespace, name);
    return axios.delete(url);
  }

  public static async get(namespace: string = "default", name: string) {
    const url = this.getLink(namespace, name);
    const { data } = await axios.get<IServiceBinding>(url);
    return data;
  }

  public static async list(namespace: string = "default") {
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
        const ns = binding.metadata.namespace;
        return axios
          .get<IK8sApiSecretResponse>(this.secretEndpoint(ns) + secretName)
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

  private static getLink(namespace?: string, name?: string): string {
    return `/api/kube/apis/servicecatalog.k8s.io/v1beta1${
      namespace ? `/namespaces/${namespace}` : ""
    }/servicebindings${name ? `/${name}` : ""}`;
  }

  private static secretEndpoint(namespace: string = "default"): string {
    return `/api/kube/api/v1/namespaces/${namespace}/secrets/`;
  }
}
