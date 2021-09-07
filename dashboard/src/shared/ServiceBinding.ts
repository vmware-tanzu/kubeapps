import { axiosWithAuth } from "./AxiosInstance";
import { APIBase } from "./Kube";
import { ICondition, ServiceCatalog } from "./ServiceCatalog";
import * as url from "./url";

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
    cluster: string,
    namespace: string,
    parameters: {},
  ) {
    const u = ServiceBinding.getLink(cluster, namespace);
    const { data } = await axiosWithAuth.post<IServiceBinding>(u, {
      apiVersion: "servicecatalog.k8s.io/v1beta1",
      kind: "ServiceBinding",
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

  public static async delete(cluster: string, namespace: string, name: string) {
    const u = this.getLink(cluster, namespace, name);
    return axiosWithAuth.delete(u);
  }

  public static async get(cluster: string, namespace: string, name: string) {
    const u = this.getLink(cluster, namespace, name);
    const { data } = await axiosWithAuth.get<IServiceBinding>(u);
    return data;
  }

  public static async list(
    cluster: string,
    namespace?: string,
  ): Promise<IServiceBindingWithSecret[]> {
    const bindings = await ServiceCatalog.getItems<IServiceBinding>(
      cluster,
      "servicebindings",
      namespace,
    );

    return Promise.all(
      bindings.map(binding => {
        const { secretName } = binding.spec;
        const ns = binding.metadata.namespace;
        return axiosWithAuth
          .get<IK8sApiSecretResponse>(url.api.k8s.secret(cluster, ns, secretName))
          .then(response => {
            return { binding, secret: response.data };
          })
          .catch(() => {
            // return with undefined secrets
            return { binding };
          });
      }),
    );
  }

  private static getLink(cluster: string, namespace?: string, name?: string): string {
    return `${APIBase(cluster)}/apis/servicecatalog.k8s.io/v1beta1${
      namespace ? `/namespaces/${namespace}` : ""
    }/servicebindings${name ? `/${name}` : ""}`;
  }
}
