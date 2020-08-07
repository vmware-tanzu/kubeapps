import { push } from "connected-react-router";
import { connect } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";

import actions from "../../actions";
import Header from "../../components/Header";
import { IStoreState } from "../../shared/types";

function mapStateToProps({
  auth: { authenticated, defaultNamespace },
  clusters,
  router: {
    location: { pathname },
  },
  config: { featureFlags, appVersion },
  catalog: { isServiceCatalogInstalled },
}: IStoreState) {
  return {
    authenticated,
    clusters,
    defaultNamespace,
    pathname,
    featureFlags,
    UI: featureFlags.ui,
    appVersion,
    isServiceCatalogInstalled,
  };
}

function mapDispatchToProps(dispatch: ThunkDispatch<IStoreState, null, Action>) {
  return {
    fetchNamespaces: (cluster: string) => dispatch(actions.namespace.fetchNamespaces(cluster)),
    createNamespace: (cluster: string, ns: string) =>
      dispatch(actions.namespace.createNamespace(cluster, ns)),
    logout: () => dispatch(actions.auth.logout()),
    push: (path: string) => dispatch(push(path)),
    setNamespace: (ns: string) => dispatch(actions.namespace.setNamespace(ns)),
    getNamespace: (cluster: string, ns: string) =>
      dispatch(actions.namespace.getNamespace(cluster, ns)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(Header);
