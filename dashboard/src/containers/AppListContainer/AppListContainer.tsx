import * as qs from "qs";
import { connect } from "react-redux";
import { RouteComponentProps } from "react-router";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";

import actions from "../../actions";
import AppList from "../../components/AppList";
import { IStoreState } from "../../shared/types";

function mapStateToProps(
  { apps, clusters: { currentCluster, clusters }, operators, config }: IStoreState,
  { location }: RouteComponentProps<{}>,
) {
  return {
    apps,
    filter: qs.parse(location.search, { ignoreQueryPrefix: true }).q || "",
    cluster: currentCluster,
    namespace: clusters[currentCluster].currentNamespace,
    customResources: operators.resources,
    isFetchingResources: operators.isFetching,
    csvs: operators.csvs,
    featureFlags: config.featureFlags,
    UI: config.featureFlags.ui,
    appVersion: config.appVersion,
  };
}

function mapDispatchToProps(dispatch: ThunkDispatch<IStoreState, null, Action>) {
  return {
    fetchAppsWithUpdateInfo: (cluster: string, ns: string, all: boolean) =>
      dispatch(actions.apps.fetchAppsWithUpdateInfo(cluster, ns, all)),
    pushSearchFilter: (filter: string) => dispatch(actions.shared.pushSearchFilter(filter)),
    getCustomResources: (cluster: string, namespace: string) =>
      dispatch(actions.operators.getResources(cluster, namespace)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(AppList);
