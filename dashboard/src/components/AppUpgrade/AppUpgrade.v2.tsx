import React, { useEffect } from "react";

import Alert from "components/js/Alert";
import { RouterAction } from "connected-react-router";
import { JSONSchema4 } from "json-schema";
import { IAppRepository, IChartState, IChartVersion, IRelease } from "../../shared/types";
import LoadingWrapper from "../LoadingWrapper/LoadingWrapper.v2";
import SelectRepoForm from "../SelectRepoForm/SelectRepoForm.v2";
import UpgradeForm from "../UpgradeForm/UpgradeForm.v2";

export interface IAppUpgradeProps {
  app?: IRelease;
  appsIsFetching: boolean;
  chartsIsFetching: boolean;
  appsError: Error | undefined;
  namespace: string;
  cluster: string;
  releaseName: string;
  repoName?: string;
  repoNamespace?: string;
  selected: IChartState["selected"];
  deployed: IChartState["deployed"];
  upgradeApp: (
    cluster: string,
    namespace: string,
    version: IChartVersion,
    chartNamespace: string,
    releaseName: string,
    values?: string,
    schema?: JSONSchema4,
  ) => Promise<boolean>;
  fetchChartVersions: (namespace: string, id: string) => Promise<IChartVersion[]>;
  getAppWithUpdateInfo: (cluster: string, namespace: string, releaseName: string) => void;
  getChartVersion: (namespace: string, id: string, chartVersion: string) => void;
  getDeployedChartVersion: (namespace: string, id: string, chartVersion: string) => void;
  push: (location: string) => RouterAction;
  // repo selector properties
  reposIsFetching: boolean;
  repoError?: Error;
  chartsError: Error | undefined;
  repo: IAppRepository;
  repos: IAppRepository[];
  checkChart: (namespace: string, repo: string, chartName: string) => any;
  fetchRepositories: (namespace: string) => void;
}

function AppUpgrade({
  app,
  appsIsFetching,
  chartsIsFetching,
  appsError,
  namespace,
  cluster,
  releaseName,
  repoName,
  repoNamespace,
  selected,
  deployed,
  upgradeApp,
  fetchChartVersions,
  getAppWithUpdateInfo,
  getChartVersion,
  getDeployedChartVersion,
  push,
  reposIsFetching,
  repoError,
  chartsError,
  repo,
  repos,
  checkChart,
  fetchRepositories,
}: IAppUpgradeProps) {
  useEffect(() => {
    getAppWithUpdateInfo(cluster, namespace, releaseName);
  }, [getAppWithUpdateInfo, cluster, namespace, releaseName]);

  const chart = app?.chart;
  useEffect(() => {
    if (
      repoName &&
      repoNamespace &&
      chart &&
      chart.metadata &&
      chart.metadata &&
      chart.metadata.name &&
      chart.metadata.version
    ) {
      const chartID = `${repoName}/${chart.metadata.name}`;
      getDeployedChartVersion(repoNamespace, chartID, chart.metadata.version);
    }
  }, [getDeployedChartVersion, app, chart, repoName, repoNamespace]);

  if (appsError) {
    return <Alert theme="danger">Found error: {appsError.message}</Alert>;
  }
  if (appsIsFetching || !app || !app.updateInfo) {
    return <LoadingWrapper loaded={false} />;
  }

  const appRepoName = repoName || app.updateInfo.repository.name;
  const repoNS = repoNamespace || app.updateInfo.repository.namespace;
  if (app && app.chart && app.chart.metadata && appRepoName) {
    return (
      <div>
        <UpgradeForm
          appCurrentVersion={app.chart.metadata.version!}
          appCurrentValues={(app.config && app.config.raw) || ""}
          chartName={app.chart.metadata.name!}
          chartsIsFetching={chartsIsFetching}
          repo={appRepoName}
          repoNamespace={repoNS}
          namespace={namespace}
          cluster={cluster}
          releaseName={releaseName}
          selected={selected}
          deployed={deployed}
          upgradeApp={upgradeApp}
          push={push}
          fetchChartVersions={fetchChartVersions}
          getChartVersion={getChartVersion}
        />
      </div>
    );
  }

  return (
    <SelectRepoForm
      isFetching={reposIsFetching}
      cluster={cluster}
      error={chartsError}
      namespace={namespace}
      repoError={repoError}
      repo={repo}
      repos={repos}
      chartName={chart?.metadata?.name!}
      checkChart={checkChart}
      fetchRepositories={fetchRepositories}
    />
  );
}

export default AppUpgrade;
