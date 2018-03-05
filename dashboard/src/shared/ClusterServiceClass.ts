import axios from "axios";
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
  public static async get(namespace?: string, name?: string): Promise<IClusterServiceClass> {
    const url = this.getLink(namespace, name);
    const { data } = await axios.get<IClusterServiceClass>(url);
    return data;
  }

  public static async list(namespace?: string): Promise<IClusterServiceClass[]> {
    const instances = await ServiceCatalog.getItems<IClusterServiceClass>("/serviceinstances");
    return instances;
  }

  private static getLink(namespace?: string, name?: string): string {
    return `/api/kube/apis/servicecatalog.k8s.io/v1beta1${
      namespace ? `/namespaces/${namespace}` : ""
    }/clusterserviceclasses${name ? `/${name}` : ""}`;
  }
}
