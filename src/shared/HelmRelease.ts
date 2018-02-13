import axios from "axios";
import { inflate } from "pako";
import { clearInterval, setInterval } from "timers";

import { hapi } from "../shared/hapi/release";
import { IApp, IChart, IHelmReleaseConfigMap } from "./types";

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

  public static async getDetails(releaseName: string) {
    const items = await this.getDetailsWithRetry(releaseName);
    // Helm/Tiller will store details in a ConfigMap for each revision,
    // so we need to filter these out to pick the latest version
    const helmConfigMap: IHelmReleaseConfigMap = items.reduce((ret: IHelmReleaseConfigMap, cm) => {
      // If the current accumulated version is higher, return it
      const curVersion = parseInt(ret.metadata.labels.VERSION, 10);
      const thisVersion = parseInt(cm.metadata.labels.VERSION, 10);
      if (curVersion > thisVersion) {
        return ret;
      }
      return cm;
    }, items[0]);
    const protoBytes = inflate(atob(helmConfigMap.data.release));
    const rel = hapi.release.Release.decode(protoBytes);
    const app: IApp = { data: rel, type: "helm" };
    return app;
  }

  private static getDetailsWithRetry(releaseName: string) {
    const getConfigMaps = () => {
      return axios.get<{ items: IHelmReleaseConfigMap[] }>(this.getConfigMapsLink([releaseName]));
    };
    return new Promise<IHelmReleaseConfigMap[]>(async (resolve, reject) => {
      let req = await getConfigMaps();
      if (req.data.items.length > 0) {
        resolve(req.data.items);
        return;
      }
      let retries = 3;
      const t = setInterval(async () => {
        if (retries <= 0) {
          clearInterval(t);
          reject();
        } else {
          req = await getConfigMaps();
          if (req.data.items.length > 0) {
            clearInterval(t);
            resolve(req.data.items);
          }
          retries = retries - 1;
        }
      }, 1000);
    });
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

  // getConfigMapsLink returns the URL for listing Helm ConfigMaps for the given
  // set of release names.
  private static getConfigMapsLink(releaseNames: string[]) {
    return `/api/kube/api/v1/namespaces/kubeapps/configmaps?labelSelector=NAME in (${releaseNames.join(
      ",",
    )})`;
  }
}
