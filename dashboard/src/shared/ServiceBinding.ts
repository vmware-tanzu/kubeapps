import { axiosWithAuth } from "./AxiosInstance";
import { APIBase } from "./Kube";
import { definedNamespaces } from "./Namespace";
import { ICondition, ServiceCatalog } from "./ServiceCatalog";

interface IK8sApiSecretResponse {
  kind: string;
  apiVersion: string;
  metadata: {
    selfLink: string;
    resourceVersion: string;
  };
  data: { [s: string]: string };
}

interface IServiceBinding {
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
    secretName: string;
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

export interface IServiceBindingWithSecret {
  binding: IServiceBinding;
  secret?: IK8sApiSecretResponse;
}

export class ServiceBinding {
  public static async create(
    bindingName: string,
    instanceRefName: string,
    namespace: string,
    parameters: {},
  ) {
    const url = ServiceBinding.getLink(namespace);
    const { data } = await axiosWithAuth.post<IServiceBinding>(url, {
      metadata: {
        name: bindingName,
      },
      spec: {
        instanceRef: {
          name: instanceRefName,
        },
        parameters,
      },
    });
    return data;
  }

  public static async delete(name: string, namespace: string) {
    const url = this.getLink(namespace, name);
    return axiosWithAuth.delete(url);
  }

  public static async get(namespace: string, name: string) {
    const url = this.getLink(namespace, name);
    const { data } = await axiosWithAuth.get<IServiceBinding>(url);
    return data;
  }

  public static async list(namespace?: string): Promise<IServiceBindingWithSecret[]> {
    const bindings = await ServiceCatalog.getItems<IServiceBinding>("servicebindings", namespace);

    return Promise.all(
      bindings.map(binding => {
        const { secretName } = binding.spec;
        const ns = binding.metadata.namespace;
        return axiosWithAuth
          .get<IK8sApiSecretResponse>(this.secretEndpoint(ns) + secretName)
          .then(response => {
            return { binding, secret: response.data };
          })
          .catch(err => {
            // return with undefined secrets
            return { binding };
          });
      }),
    );
  }

  private static getLink(namespace?: string, name?: string): string {
    return `${APIBase}/apis/servicecatalog.k8s.io/v1beta1${
      namespace ? `/namespaces/${namespace}` : ""
    }/servicebindings${name ? `/${name}` : ""}`;
  }

  private static secretEndpoint(namespace: string = definedNamespaces.default): string {
    return `${APIBase}/api/v1/namespaces/${namespace}/secrets/`;
  }
}
