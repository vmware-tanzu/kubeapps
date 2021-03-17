import { useEffect } from "react";

import Alert from "components/js/Alert";
import { RouterAction } from "connected-react-router";
import { JSONSchema4 } from "json-schema";
import {
  FetchError,
  IAppRepository,
  IChartState,
  IChartVersion,
  IRelease,
  UpgradeError,
} from "../../shared/types";
import LoadingWrapper from "../LoadingWrapper/LoadingWrapper";
import SelectRepoForm from "../SelectRepoForm/SelectRepoForm";
import UpgradeForm from "../UpgradeForm/UpgradeForm";

export interface IAppUpgradeProps {
  app?: IRelease;
  appsIsFetching: boolean;
  chartsIsFetching: boolean;
  error?: FetchError | UpgradeError;
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
  fetchChartVersions: (cluster: string, namespace: string, id: string) => Promise<IChartVersion[]>;
  getAppWithUpdateInfo: (cluster: string, namespace: string, releaseName: string) => void;
  getChartVersion: (cluster: string, namespace: string, id: string, chartVersion: string) => void;
  getDeployedChartVersion: (
    cluster: string,
    namespace: string,
    id: string,
    chartVersion: string,
  ) => void;
  push: (location: string) => RouterAction;
  // repo selector properties
  reposIsFetching: boolean;
  repoError?: Error;
  chartsError: Error | undefined;
  repo: IAppRepository;
  repos: IAppRepository[];
  checkChart: (cluster: string, namespace: string, repo: string, chartName: string) => any;
  fetchRepositories: (namespace: string) => void;
}

function AppUpgrade({
  app,
  appsIsFetching,
  chartsIsFetching,
  error,
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
      getDeployedChartVersion(cluster, repoNamespace, chartID, chart.metadata.version);
    }
  }, [getDeployedChartVersion, app, chart, repoName, repoNamespace, cluster]);

  if (error && error.constructor === FetchError) {
    return <Alert theme="danger">Unable to retrieve the current app: {error.message}</Alert>;
  }

  if (appsIsFetching || !app || !app.updateInfo) {
    return (
      <LoadingWrapper
        loadingText={`Fetching ${releaseName}...`}
        className="margin-t-xxl"
        loaded={false}
      />
    );
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
          error={error}
          fetchChartVersions={fetchChartVersions}
          getChartVersion={getChartVersion}
        />
      </div>
    );
  }
  /* eslint-disable @typescript-eslint/no-non-null-asserted-optional-chain */
  return (
    <SelectRepoForm cluster={cluster} namespace={namespace} chartName={chart?.metadata?.name!} />
  );
}

export default AppUpgrade;
