import { goBack, push } from "connected-react-router";
import { connect } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";

import actions from "../../actions";

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
    isFetching: apps.isFetching || repos.isFetching,
    error: apps.error || charts.selected.error,
    kubeappsNamespace: config.namespace,
    namespace: params.namespace,
    releaseName: params.releaseName,
    repo: repos.repo,
    repoError: repos.errors.fetch,
    repos: repos.repos,
    selected: charts.selected,
  };
}

function mapDispatchToProps(dispatch: ThunkDispatch<IStoreState, null, Action>) {
  return {
    checkChart: (repo: string, chartName: string) =>
      dispatch(actions.repos.checkChart(repo, chartName)),
    clearRepo: () => dispatch(actions.repos.clearRepo()),
    fetchChartVersions: (id: string) => dispatch(actions.charts.fetchChartVersions(id)),
    fetchRepositories: () => dispatch(actions.repos.fetchRepos()),
    getAppWithUpdateInfo: (releaseName: string, ns: string) =>
      dispatch(actions.apps.getAppWithUpdateInfo(releaseName, ns)),
    getChartValues: (id: string, version: string) =>
      dispatch(actions.charts.getChartValues(id, version)),
    getChartVersion: (id: string, version: string) =>
      dispatch(actions.charts.getChartVersion(id, version)),
    push: (location: string) => dispatch(push(location)),
    goBack: () => dispatch(goBack()),
    upgradeApp: (version: IChartVersion, releaseName: string, namespace: string, values?: string) =>
      dispatch(actions.apps.upgradeApp(version, releaseName, namespace, values)),
  };
}

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(AppUpgrade);
