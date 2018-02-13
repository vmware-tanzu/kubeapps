import { inflate } from "pako";
import { Dispatch } from "redux";
import { createAction, getReturnOfExpression } from "typesafe-actions";

import { hapi } from "../shared/hapi/release";
import { HelmRelease } from "../shared/HelmRelease";
import { IApp, IHelmRelease, IHelmReleaseConfigMap, IStoreState } from "../shared/types";
import * as url from "../shared/url";

export const requestApps = createAction("REQUEST_APPS");
export const receiveApps = createAction("RECEIVE_APPS", (apps: IApp[]) => {
  return {
    apps,
    type: "RECEIVE_APPS",
  };
});
export const selectApp = createAction("SELECT_APP", (app: IApp) => {
  return {
    app,
    type: "SELECT_APP",
  };
});

const allActions = [requestApps, receiveApps, selectApp].map(getReturnOfExpression);
export type AppsAction = typeof allActions[number];

export function getApp(releaseName: string) {
  return async (dispatch: Dispatch<IStoreState>): Promise<void> => {
    const app = await HelmRelease.getDetails(releaseName);
    dispatch(selectApp(app));
  };
}

export function fetchApps() {
  return (dispatch: Dispatch<IStoreState>): Promise<{}> => {
    dispatch(requestApps());
    return fetch(url.api.helmreleases.list())
      .then(response => response.json())
      .then((json: { items: IHelmRelease[] }) => {
        // fetch the ConfigMap for each HelmRelease object
        // const releasesByName: { [s: string]: IHelmRelease } = json.items.reduce((acc, hr) => {
        //   acc[hr.metadata.name] = hr;
        //   return acc;
        // }, {});
        const releaseNames = json.items.map(hr => {
          return `${hr.metadata.namespace}-${hr.metadata.name}`;
        }, {});
        return fetch(url.api.helmreleases.listDetails(releaseNames))
          .then(response => response.json())
          .then((details: { items: IHelmReleaseConfigMap[] }) => {
            // Helm/Tiller will store details in a ConfigMap for each revision,
            // so we need to filter these out to pick the latest version
            const cms: { [s: string]: IHelmReleaseConfigMap } = details.items.reduce((acc, cm) => {
              const releaseName = cm.metadata.labels.NAME;
              // If we've already found a version for this release, only
              // replace it if the version is greater
              if (releaseName in acc) {
                const curVersion = parseInt(acc[releaseName].metadata.labels.VERSION, 10);
                const thisVersion = parseInt(cm.metadata.labels.VERSION, 10);
                if (curVersion > thisVersion) {
                  return acc;
                }
              }
              acc[releaseName] = cm;
              return acc;
            }, {});
            // Iterate through ConfigMaps to decode base64, ungzip (inflate) and
            // parse as a protobuf message
            const apps: IApp[] = [];
            for (const key of Object.keys(cms)) {
              const cm = cms[key];
              const protoBytes = inflate(atob(cm.data.release));
              const rel = hapi.release.Release.decode(protoBytes);
              // const helmrelease = releasesByName[key];
              const app: IApp = { data: rel, type: "helm" };
              // const repoName =
              //   helmrelease.metadata.annotations["apprepositories.kubeapps.com/repo-name"];
              // if (repoName) {
              //   app.repo = {
              //     name: repoName,
              //     url: helmrelease.spec.repoUrl,
              //   };
              // }
              apps.push(app);
            }
            dispatch(receiveApps(apps));
            return apps;
          });
      });
  };
}
