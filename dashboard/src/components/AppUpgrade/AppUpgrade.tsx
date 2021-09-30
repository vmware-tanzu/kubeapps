import actions from "actions";
import Alert from "components/js/Alert";
import {
  AvailablePackageReference,
  InstalledPackageDetail,
  InstalledPackageReference,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import * as ReactRouter from "react-router";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import { FetchError, IAppRepository, IChartState, IStoreState, UpgradeError } from "shared/types";
import LoadingWrapper from "../LoadingWrapper/LoadingWrapper";
import SelectRepoForm from "../SelectRepoForm/SelectRepoForm";
import UpgradeForm from "../UpgradeForm/UpgradeForm";

export interface IAppUpgradeProps {
  app?: InstalledPackageDetail;
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
  pluginName: string;
  pluginVersion: string;
}

function AppUpgrade() {
  const dispatch: ThunkDispatch<IStoreState, null, Action> = useDispatch();
  const { cluster, namespace, releaseName, pluginName, pluginVersion } =
    ReactRouter.useParams() as IRouteParams;
  const {
    apps: { selected: app, isFetching: appsIsFetching, error },
    charts: { isFetching: chartsIsFetching, selected, deployed },
    repos: { repo },
  } = useSelector((state: IStoreState) => state);

  const repoName = repo?.metadata?.name || app?.availablePackageRef?.context?.namespace;
  const repoNamespace = repo?.metadata?.namespace || app?.availablePackageRef?.context?.namespace;

  const [pluginObj] = useState(
    selected.availablePackageDetail?.availablePackageRef?.plugin ??
      ({ name: pluginName, version: pluginVersion } as Plugin),
  );

  useEffect(() => {
    dispatch(
      actions.apps.getApp({
        context: { cluster: cluster, namespace: namespace },
        identifier: releaseName,
        plugin: pluginObj,
      } as InstalledPackageReference),
    );
  }, [dispatch, cluster, namespace, releaseName, pluginObj]);

  useEffect(() => {
    dispatch(
      actions.charts.getDeployedChartVersion(
        {
          context: {
            cluster: app?.availablePackageRef?.context?.cluster ?? cluster,
            namespace: repoNamespace ?? "",
          },
          identifier: app?.availablePackageRef?.identifier ?? "",
          plugin: app?.availablePackageRef?.plugin,
        } as AvailablePackageReference,
        app?.currentVersion?.pkgVersion,
      ),
    );
  }, [dispatch, app, repoName, repoNamespace, cluster]);

  if (error && error.constructor === FetchError) {
    return <Alert theme="danger">Unable to retrieve the current app: {error.message}</Alert>;
  }

  if (appsIsFetching || !app) {
    return (
      <LoadingWrapper
        loadingText={`Fetching ${releaseName}...`}
        className="margin-t-xxl"
        loaded={false}
      />
    );
  }
  if (
    app?.currentVersion?.pkgVersion &&
    app?.valuesApplied &&
    app?.availablePackageRef?.identifier &&
    repoNamespace &&
    namespace &&
    cluster &&
    releaseName &&
    selected &&
    deployed
  ) {
    return (
      <div>
        <UpgradeForm
          appCurrentVersion={app.currentVersion.pkgVersion}
          appCurrentValues={app.valuesApplied}
          packageId={app.availablePackageRef.identifier}
          chartsIsFetching={chartsIsFetching}
          repoNamespace={repoNamespace}
          namespace={namespace}
          cluster={cluster}
          releaseName={releaseName}
          selected={selected}
          deployed={deployed}
          error={error}
          plugin={pluginObj}
        />
      </div>
    );
  }
  return <SelectRepoForm cluster={cluster} namespace={namespace} app={app} />;
}

export default AppUpgrade;
