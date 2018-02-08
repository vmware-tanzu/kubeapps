import axios from "axios";
import { IChart } from "./types";

export class HelmRelease {
  public static async create(chart: IChart, releaseName: string, namespace: string) {
    const endpoint = HelmRelease.getResourceLink(namespace);
    const { data } = await axios.post(endpoint, {
      data: {
        apiVersion: "helm.bitnami.com/v1",
        kind: "HelmRelease",
        metadata: {
          releaseName,
        },
        spec: {
          chartName: chart.attributes.name,
          repoUrl: chart.attributes.repo.url,
          version: chart.relationships.latestChartVersion.data.version,
        },
      },
    });
    return data;
  }

  public static async delete(selfLink: string) {
    const { data } = await axios.delete(selfLink);
    return data;
  }

  // private static getSelfLink(name: string, namespace: string) {
  //   return `/api/kube/apis/helm.bitnami.com/v1/namespaces/${namespace}/helmreleases/${name}`;
  // }

  private static getResourceLink(namespace?: string) {
    if (namespace) {
      return `/api/kube/apis/helm.bitnami.com/v1/namespaces/${namespace}/helmreleases`;
    } else {
      return `/api/kube/apis/helm.bitnami.com/v1/helmreleases`;
    }
  }
}
