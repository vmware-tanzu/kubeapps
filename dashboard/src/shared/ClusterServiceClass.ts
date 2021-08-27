import { axiosWithAuth } from "./AxiosInstance";
import { APIBase } from "./Kube";
import { ServiceCatalog } from "./ServiceCatalog";

export interface IClusterServiceClass {
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

export class ClusterServiceClass {
  public static async get(
    cluster: string,
    namespace?: string,
    name?: string,
  ): Promise<IClusterServiceClass> {
    const url = this.getLink(cluster, namespace, name);
    const { data } = await axiosWithAuth.get<IClusterServiceClass>(url);
    return data;
  }

  public static async list(cluster: string): Promise<IClusterServiceClass[]> {
    const instances = await ServiceCatalog.getItems<IClusterServiceClass>(
      cluster,
      "serviceinstances",
    );
    return instances;
  }

  private static getLink(cluster: string, namespace?: string, name?: string): string {
    return `${APIBase(cluster)}/apis/servicecatalog.k8s.io/v1beta1${
      namespace ? `/namespaces/${namespace}` : ""
    }/clusterserviceclasses${name ? `/${name}` : ""}`;
  }
}
