import { axios } from "./Auth";
import { IAppConfigMap } from "./types";

export class App {
  public static async waitForDeletion(name: string) {
    const timeout = 30000; // 30s
    return new Promise((resolve, reject) => {
      const interval = setInterval(async () => {
        const { data: { items: allConfigMaps } } = await axios.get<{
          items: IAppConfigMap[];
        }>(this.getConfigMapsLink([name]));
        if (allConfigMaps.length === 0) {
          clearInterval(interval);
          resolve();
        }
      }, 500);
      setTimeout(() => {
        clearInterval(interval);
        reject(`Timeout after ${timeout / 1000} seconds`);
      }, timeout);
    });
  }

  public static async appExists(releaseName: string) {
    const { data: { items: allConfigMaps } } = await axios.get<{
      items: IAppConfigMap[];
    }>(this.getConfigMapsLink([releaseName]));
    if (allConfigMaps.length === 0) {
      return false;
    }
    return true;
  }

  // getConfigMapsLink returns the URL for listing Helm ConfigMaps for the given
  // set of release names.
  public static getConfigMapsLink(releaseNames?: string[]) {
    let query = "";
    if (releaseNames) {
      query = `,NAME in (${releaseNames.join(",")})`;
    }
    return `/api/kube/api/v1/namespaces/kubeapps/configmaps?labelSelector=OWNER=TILLER${query}`;
  }
}
