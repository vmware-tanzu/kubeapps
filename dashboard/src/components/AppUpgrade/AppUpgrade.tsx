import actions from "actions";
import Alert from "components/js/Alert";
import { useEffect } from "react";
import { useDispatch, useSelector } from "react-redux";
import * as ReactRouter from "react-router";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import {
  FetchError,
  IAppRepository,
  IChartState,
  IRelease,
  IStoreState,
  UpgradeError,
} from "shared/types";
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
  reposIsFetching: boolean;
  repoError?: Error;
  chartsError: Error | undefined;
  repo: IAppRepository;
  repos: IAppRepository[];
}

interface IRouteParams {
  cluster: string;
  namespace: string;
  releaseName: string;
}

function AppUpgrade() {
  const dispatch: ThunkDispatch<IStoreState, null, Action> = useDispatch();
  const { cluster, namespace, releaseName } = ReactRouter.useParams() as IRouteParams;
  const {
    apps: { selected: app, isFetching: appsIsFetching, error },
    charts: { isFetching: chartsIsFetching, selected, deployed },
    repos: { repo },
  } = useSelector((state: IStoreState) => state);

  const repoName = repo?.metadata?.name || app?.updateInfo?.repository?.namespace;
  const repoNamespace = repo?.metadata?.namespace || app?.updateInfo?.repository?.namespace;

  useEffect(() => {
    dispatch(actions.apps.getAppWithUpdateInfo(cluster, namespace, releaseName));
  }, [dispatch, cluster, namespace, releaseName]);

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
      dispatch(
        actions.charts.getDeployedChartVersion(
          cluster,
          repoNamespace,
          chartID,
          chart.metadata.version,
        ),
      );
    }
  }, [dispatch, app, chart, repoName, repoNamespace, cluster]);

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
          error={error}
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
