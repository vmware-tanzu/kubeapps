import * as qs from "qs";
import { connect } from "react-redux";
import { RouteComponentProps } from "react-router";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";

import actions from "../../actions";
import AppList from "../../components/AppList";
import { IStoreState } from "../../shared/types";

function mapStateToProps(
  { apps, namespace, charts }: IStoreState,
  { location }: RouteComponentProps<{}>,
) {
  return {
    apps,
    filter: qs.parse(location.search, { ignoreQueryPrefix: true }).q || "",
    namespace: namespace.current,
  };
}

function mapDispatchToProps(dispatch: ThunkDispatch<IStoreState, null, Action>) {
  return {
    fetchAppsWithUpdateInfo: (ns: string, all: boolean) =>
      dispatch(actions.apps.fetchAppsWithUpdateInfo(ns, all)),
    pushSearchFilter: (filter: string) => dispatch(actions.shared.pushSearchFilter(filter)),
  };
}

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(AppList);
