import { inflate } from "pako";
import { clearInterval, setInterval } from "timers";

import { App } from "./App";
import { AppRepository } from "./AppRepository";
import { axios } from "./Auth";
import { hapi } from "./hapi/release";
import { IApp, IAppConfigMap, IChart, IChartVersion, IHelmRelease } from "./types";
import * as url from "./url";

export class HelmRelease {
  public static async create(
    releaseName: string,
    namespace: string,
    chartVersion: IChartVersion,
    values?: string,
  ) {
    const chartAttrs = chartVersion.relationships.chart.data;
    const repo = await AppRepository.get(chartAttrs.repo.name);
    const auth = repo.spec.auth;
    const endpoint = HelmRelease.getResourceLink(namespace);
    const { data } = await axios.post(endpoint, {
      apiVersion: "helm.bitnami.com/v1",
      kind: "HelmRelease",
      metadata: {
        annotations: {
          "apprepositories.kubeapps.com/repo-name": chartAttrs.repo.name,
        },
        name: releaseName,
      },
      spec: {
        auth,
        chartName: chartAttrs.name,
        releaseName,
        repoUrl: chartAttrs.repo.url,
        values,
        version: chartVersion.attributes.version,
      },
    });
    return data;
  }

  public static async upgrade(
    releaseName: string,
    namespace: string,
    chartVersion: IChartVersion,
    values?: string,
  ) {
    const chartAttrs = chartVersion.relationships.chart.data;
    const repo = await AppRepository.get(chartAttrs.repo.name);
    const auth = repo.spec.auth;
    const endpoint = HelmRelease.getSelfLink(releaseName, namespace);
    const { data } = await axios.patch(
      endpoint,
      {
        apiVersion: "helm.bitnami.com/v1",
        kind: "HelmRelease",
        metadata: {
          name: releaseName,
        },
        spec: {
          auth,
          chartName: chartAttrs.name,
          releaseName,
          repoUrl: chartAttrs.repo.url,
          values,
          version: chartVersion.attributes.version,
        },
      },
      {
        headers: { "Content-Type": "application/merge-patch+json" },
      },
    );
    return data;
  }

  public static async delete(releaseName: string, namespace: string) {
    // strip namespace from release name
    const { data } = await axios.delete(this.getSelfLink(releaseName, namespace));
    return data;
  }

  public static async getAllHelmReleases(namespace?: string) {
    const { data: { items: helmReleaseList } } = await axios.get<{ items: IHelmRelease[] }>(
      this.getResourceLink(namespace),
    );
    return helmReleaseList;
  }

  public static async getHelmRelease(releaseName: string, namespace: string) {
    const helmReleaseList = await this.getAllHelmReleases(namespace);
    let helmRelease = "";
    helmReleaseList.forEach(r => {
      if (r.spec.releaseName === releaseName) {
        helmRelease = r.metadata.name;
      }
    });
    return helmRelease;
  }

  public static async getAllWithDetails(namespace?: string) {
    const helmReleaseList = await this.getAllHelmReleases(namespace);
    // Convert list of HelmReleases to release name -> HelmRelease pair
    const helmReleaseMap = helmReleaseList.reduce((acc, hr) => {
      const releaseName =
        !hr.spec.releaseName || hr.spec.releaseName === ""
          ? `${hr.metadata.name}-${hr.metadata.namespace}`
          : hr.spec.releaseName;
      acc[releaseName] = hr;
      return acc;
    }, new Map<string, IHelmRelease>());

    // Get the HelmReleaseConfigMaps for all HelmReleases
    const { data: { items: allConfigMaps } } = await axios.get<{
      items: IAppConfigMap[];
    }>(App.getConfigMapsLink());

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
    }, new Map<string, IAppConfigMap>());

    // Go through all HelmReleaseConfigMaps and parse as IApp objects
    const releases = Object.keys(cms)
      .map(key => this.parseRelease(cms[key], helmReleaseMap[key]))
      // Exclude releases that are not of this namespace
      .filter(app => !namespace || app.data.namespace === namespace);

    // Fetch charts for each app
    return await Promise.all<IApp>(releases.map(rel => this.getChart(rel)));
  }

  public static async getDetails(releaseName: string, namespace: string) {
    let hr;
    try {
      const i = await axios.get<IHelmRelease>(this.getSelfLink(releaseName, namespace));
      hr = i.data;
    } catch (e) {
      // HelmRelease not available
    }
    const items = await this.getDetailsWithRetry(releaseName);
    // Helm/Tiller will store details in a ConfigMap for each revision,
    // so we need to filter these out to pick the latest version
    const helmConfigMap: IAppConfigMap = items.reduce((ret, cm) => {
      return this.getNewest(ret, cm);
    }, items[0]);

    const app = this.parseRelease(helmConfigMap, hr);
    return await this.getChart(app);
  }

  private static getDetailsWithRetry(releaseName: string) {
    const getConfigMaps = () => {
      return axios.get<{ items: IAppConfigMap[] }>(App.getConfigMapsLink([releaseName]));
    };
    return new Promise<IAppConfigMap[]>(async (resolve, reject) => {
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

  // Takes two IAppConfigMaps and returns the highest version
  private static getNewest(cm1: IAppConfigMap, cm2: IAppConfigMap) {
    const cm1Version = parseInt(cm1.metadata.labels.VERSION, 10);
    const cm2Version = parseInt(cm2.metadata.labels.VERSION, 10);
    return cm1Version > cm2Version ? cm1 : cm2;
  }

  // decode base64, ungzip (inflate) and parse as a protobuf message
  private static parseRelease(cm: IAppConfigMap, hr?: IHelmRelease): IApp {
    const protoBytes = inflate(atob(cm.data.release));
    const rel = hapi.release.Release.decode(protoBytes);
    const app: IApp = { data: rel, type: "helm", hr };
    if (hr && hr.metadata) {
      const repoName = hr.metadata.annotations["apprepositories.kubeapps.com/repo-name"];
      if (repoName) {
        app.repo = {
          name: repoName,
          url: hr.spec.repoUrl,
        };
      }
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
