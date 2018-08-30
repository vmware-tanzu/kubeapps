import * as qs from "qs";
import { connect } from "react-redux";
import { RouteComponentProps } from "react-router";
import { Dispatch } from "redux";

import actions from "../../actions";
import AppList from "../../components/AppList";
import { IStoreState } from "../../shared/types";

function mapStateToProps({ apps, namespace }: IStoreState, { location }: RouteComponentProps<{}>) {
  return {
    apps,
    filter: qs.parse(location.search, { ignoreQueryPrefix: true }).q || "",
    namespace: namespace.current,
  };
}

function mapDispatchToProps(dispatch: Dispatch<IStoreState>) {
  return {
    fetchApps: (ns: string, all: boolean) => dispatch(actions.apps.fetchApps(ns, all)),
    pushSearchFilter: (filter: string) => dispatch(actions.shared.pushSearchFilter(filter)),
    toggleListAll: () => dispatch(actions.apps.toggleListAll()),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(AppList);
