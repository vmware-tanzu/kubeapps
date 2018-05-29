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
      tillerReleaseName: string;
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
    repos: repos.repos,
    tillerReleaseName: params.tillerReleaseName,
  };
}

function mapDispatchToProps(dispatch: Dispatch<IStoreState>) {
  return {
    deployChart: (
      helmCRDReleaseName: string,
      version: IChartVersion,
      tillerReleaseName: string,
      namespace: string,
      values?: string,
      resourceVersion?: string,
    ) =>
      dispatch(
        actions.apps.deployChart(
          helmCRDReleaseName,
          version,
          tillerReleaseName,
          namespace,
          values,
          resourceVersion,
        ),
      ),
    fetchRepositories: () => dispatch(actions.repos.fetchRepos()),
    getApp: (tillerReleaseName: string, ns: string) =>
      dispatch(actions.apps.getApp("", tillerReleaseName, ns)),
    push: (location: string) => dispatch(push(location)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(AppMigrate);
