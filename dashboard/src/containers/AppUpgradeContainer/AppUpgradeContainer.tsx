import { goBack, push } from "connected-react-router";
import { connect } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import actions from "../../actions";
import AppUpgrade from "../../components/AppUpgrade";
import { IChartVersion, IStoreState } from "../../shared/types";
import { JSONSchemaType } from "ajv";

interface IRouteProps {
  match: {
    params: {
      cluster: string;
      namespace: string;
      releaseName: string;
    };
  };
}

function mapStateToProps(
  { apps, charts, config, repos }: IStoreState,
  { match: { params } }: IRouteProps,
) {
  return {
    app: apps.selected,
    appsIsFetching: apps.isFetching,
    chartsIsFetching: charts.isFetching,
    reposIsFetching: repos.isFetching,
    error: apps.error,
    chartsError: charts.selected.error,
    kubeappsNamespace: config.kubeappsNamespace,
    namespace: params.namespace,
    cluster: params.cluster,
    releaseName: params.releaseName,
    repo: repos.repo,
    repoError: repos.errors.fetch,
    repos: repos.repos,
    selected: charts.selected,
    deployed: charts.deployed,
    repoName:
      (repos.repo.metadata && repos.repo.metadata.name) ||
      (apps.selected && apps.selected.updateInfo && apps.selected.updateInfo.repository.name),
    repoNamespace:
      (repos.repo.metadata && repos.repo.metadata.namespace) ||
      (apps.selected && apps.selected.updateInfo && apps.selected.updateInfo.repository.namespace),
  };
}

function mapDispatchToProps(dispatch: ThunkDispatch<IStoreState, null, Action>) {
  return {
    checkChart: (cluster: string, namespace: string, repo: string, chartName: string) =>
      dispatch(actions.repos.checkChart(cluster, namespace, repo, chartName)),
    clearRepo: () => dispatch(actions.repos.clearRepo()),
    fetchChartVersions: (cluster: string, namespace: string, id: string) =>
      dispatch(actions.charts.fetchChartVersions(cluster, namespace, id)),
    fetchRepositories: (namespace: string) => dispatch(actions.repos.fetchRepos(namespace)),
    getAppWithUpdateInfo: (cluster: string, namespace: string, releaseName: string) =>
      dispatch(actions.apps.getAppWithUpdateInfo(cluster, namespace, releaseName)),
    getChartVersion: (cluster: string, namespace: string, id: string, version: string) =>
      dispatch(actions.charts.getChartVersion(cluster, namespace, id, version)),
    push: (location: string) => dispatch(push(location)),
    goBack: () => dispatch(goBack()),
    upgradeApp: (
      cluster: string,
      namespace: string,
      version: IChartVersion,
      chartNamespace: string,
      releaseName: string,
      values?: string,
      schema?: JSONSchemaType<any>,
    ) =>
      dispatch(
        actions.apps.upgradeApp(
          cluster,
          namespace,
          version,
          chartNamespace,
          releaseName,
          values,
          schema,
        ),
      ),
    getDeployedChartVersion: (cluster: string, namespace: string, id: string, version: string) =>
      dispatch(actions.charts.getDeployedChartVersion(cluster, namespace, id, version)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(AppUpgrade);
