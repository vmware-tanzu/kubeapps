import { connect } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";

import actions from "../../actions";
import RollbackButton from "../../components/AppView/AppControls/RollbackButton";
import { IChartVersion, IStoreState } from "../../shared/types";

interface IButtonProps {
  namespace: string;
  releaseName: string;
}

function mapStateToProps({ apps, charts, repos, config }: IStoreState, props: IButtonProps) {
  return {
    app: apps.selected!,
    error: apps.error || repos.errors.fetch,
    namespace: props.namespace,
    releaseName: props.releaseName,
    chartVersion: charts.selected.version,
    loading: charts.isFetching || apps.isFetching || repos.isFetching,
    repo: repos.repo,
    repoError: repos.errors.fetch,
    repos: repos.repos,
    kubeappsNamespace: config.namespace,
  };
}

function mapDispatchToProps(dispatch: ThunkDispatch<IStoreState, null, Action>) {
  return {
    rollbackApp: (
      chartVersion: IChartVersion,
      releaseName: string,
      revision: number,
      namespace: string,
      values: string,
    ) => dispatch(actions.apps.rollbackApp(chartVersion, releaseName, revision, namespace, values)),
    getChartVersion: (id: string, version: string) =>
      dispatch(actions.charts.getChartVersion(id, version)),
    checkChart: (repo: string, chartName: string) =>
      dispatch(actions.repos.checkChart(repo, chartName)),
    fetchRepositories: () => dispatch(actions.repos.fetchRepos()),
  };
}

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(RollbackButton);
