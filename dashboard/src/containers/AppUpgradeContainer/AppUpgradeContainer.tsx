import { goBack, push } from "connected-react-router";
import { connect } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";

import actions from "../../actions";

import { JSONSchema4 } from "json-schema";
import AppUpgrade from "../../components/AppUpgrade";
import { IChartVersion, IStoreState } from "../../shared/types";

interface IRouteProps {
  match: {
    params: {
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
    reposIsFetching: repos.isFetching,
    appsError: apps.error,
    chartsError: charts.selected.error,
    kubeappsNamespace: config.namespace,
    namespace: params.namespace,
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
    checkChart: (namespace: string, repo: string, chartName: string) =>
      dispatch(actions.repos.checkChart(namespace, repo, chartName)),
    clearRepo: () => dispatch(actions.repos.clearRepo()),
    fetchChartVersions: (namespace: string, id: string) => dispatch(actions.charts.fetchChartVersions(namespace, id)),
    fetchRepositories: (namespace: string) => dispatch(actions.repos.fetchRepos(namespace)),
    getAppWithUpdateInfo: (namespace: string, releaseName: string) =>
      dispatch(actions.apps.getAppWithUpdateInfo(namespace, releaseName)),
    getChartVersion: (namespace: string, id: string, version: string) =>
      dispatch(actions.charts.getChartVersion(namespace, id, version)),
    push: (location: string) => dispatch(push(location)),
    goBack: () => dispatch(goBack()),
    upgradeApp: (
      version: IChartVersion,
      releaseName: string,
      namespace: string,
      values?: string,
      schema?: JSONSchema4,
    ) => dispatch(actions.apps.upgradeApp(version, releaseName, namespace, values, schema)),
    getDeployedChartVersion: (namespace: string, id: string, version: string) =>
      dispatch(actions.charts.getDeployedChartVersion(namespace, id, version)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(AppUpgrade);
