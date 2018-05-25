import { connect } from "react-redux";
import { push } from "react-router-redux";
import { Dispatch } from "redux";

import actions from "../../actions";
import AppMigrate from "../../components/AppMigrate";
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
  { repos, apps, catalog, charts }: IStoreState,
  { match: { params } }: IRouteProps,
) {
  return {
    app: apps.selected,
    error: apps.error,
    namespace: params.namespace,
    releaseName: params.releaseName,
    repos: repos.repos,
  };
}

function mapDispatchToProps(dispatch: Dispatch<IStoreState>) {
  return {
    deployChart: (
      hrName: string,
      version: IChartVersion,
      releaseName: string,
      namespace: string,
      values?: string,
      resourceVersion?: string,
    ) =>
      dispatch(
        actions.apps.deployChart(hrName, version, releaseName, namespace, values, resourceVersion),
      ),
    fetchRepositories: () => dispatch(actions.repos.fetchRepos()),
    getApp: (releaseName: string, ns: string) => dispatch(actions.apps.getApp("", releaseName, ns)),
    push: (location: string) => dispatch(push(location)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(AppMigrate);
