import axios from "axios";
import { inflate } from "pako";
import { clearInterval, setInterval } from "timers";

import { hapi } from "../shared/hapi/release";
import { IApp, IChart, IHelmRelease, IHelmReleaseConfigMap } from "./types";
import * as url from "./url";

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

  public static async delete(releaseName: string, namespace: string) {
    // strip namespace from release name
    const hrName = releaseName.replace(new RegExp(`^${namespace}-`), "");
    const { data } = await axios.delete(this.getSelfLink(hrName, namespace));
    return data;
  }

  public static async getAllWithDetails() {
    const { data: { items: helmReleaseList } } = await axios.get<{ items: IHelmRelease[] }>(
      this.getResourceLink(),
    );
    // Convert list of HelmReleases to release name -> HelmRelease pair
    const helmReleaseMap = helmReleaseList.reduce((acc, hr) => {
      acc[`${hr.metadata.namespace}-${hr.metadata.name}`] = hr;
      return acc;
    }, new Map<string, IHelmRelease>());

    // Get the HelmReleaseConfigMaps for all HelmReleases
    const { data: { items: allConfigMaps } } = await axios.get<{ items: IHelmReleaseConfigMap[] }>(
      this.getConfigMapsLink(Object.keys(helmReleaseMap)),
    );

    // Convert list of HelmReleaseConfigMaps to release name -> latest
    // HelmReleaseConfigMap pair
    const cms = allConfigMaps.reduce((acc, cm) => {
      const releaseName = cm.metadata.labels.NAME;
      // If we've already found a version for this release, only
      // replace it if the version is greater
      if (releaseName in acc) {
        acc[releaseName] = this.getNewest(acc[releaseName], cm);
      } else {
        acc[releaseName] = cm;
      }
      return acc;
    }, new Map<string, IHelmReleaseConfigMap>());

    // Go through all HelmReleaseConfigMaps and parse as IApp objects
    const apps = Object.keys(cms).map(key => this.parseRelease(helmReleaseMap[key], cms[key]));

    // Fetch charts for each app
    return Promise.all<IApp>(apps.map(async app => this.getChart(app)));
  }

  public static async getDetails(releaseName: string, namespace: string) {
    // strip namespace from release name
    const hrName = releaseName.replace(new RegExp(`^${namespace}-`), "");
    const { data: hr } = await axios.get<IHelmRelease>(this.getSelfLink(hrName, namespace));
    const items = await this.getDetailsWithRetry(releaseName);
    // Helm/Tiller will store details in a ConfigMap for each revision,
    // so we need to filter these out to pick the latest version
    const helmConfigMap: IHelmReleaseConfigMap = items.reduce((ret, cm) => {
      return this.getNewest(ret, cm);
    }, items[0]);

    const app = this.parseRelease(hr, helmConfigMap);
    return await this.getChart(app);
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

  private static getSelfLink(name: string, namespace: string) {
    return `/api/kube/apis/helm.bitnami.com/v1/namespaces/${namespace}/helmreleases/${name}`;
  }

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

  // Takes two IHelmReleaseConfigMaps and returns the highest version
  private static getNewest(cm1: IHelmReleaseConfigMap, cm2: IHelmReleaseConfigMap) {
    const cm1Version = parseInt(cm1.metadata.labels.VERSION, 10);
    const cm2Version = parseInt(cm2.metadata.labels.VERSION, 10);
    return cm1Version > cm2Version ? cm1 : cm2;
  }

  // decode base64, ungzip (inflate) and parse as a protobuf message
  private static parseRelease(hr: IHelmRelease, cm: IHelmReleaseConfigMap): IApp {
    const protoBytes = inflate(atob(cm.data.release));
    const rel = hapi.release.Release.decode(protoBytes);
    const app: IApp = { data: rel, type: "helm" };
    const repoName = hr.metadata.annotations["apprepositories.kubeapps.com/repo-name"];
    if (repoName) {
      app.repo = {
        name: repoName,
        url: hr.spec.repoUrl,
      };
    }
    return app;
  }

  private static async getChart(app: IApp) {
    if (app.repo && app.data.chart && app.data.chart.metadata) {
      const res = await axios.get<{ data: IChart }>(
        url.api.charts.get(`${app.repo.name}/${app.data.chart.metadata.name}`),
      );
      app.chart = res.data.data;
    }
    return app;
  }
}
