import { push } from "connected-react-router";
import { connect } from "react-redux";
import { RouteComponentProps } from "react-router";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";

import actions from "../../actions";
import Header from "../../components/Header";
import { IStoreState } from "../../shared/types";

interface IState extends IStoreState {
  router: RouteComponentProps<{}>;
}

function mapStateToProps({
  auth: { authenticated, defaultNamespace },
  clusters,
  router: {
    location: { pathname },
  },
  config: { featureFlags },
}: IState) {
  return {
    authenticated,
    cluster: clusters.clusters.default,
    defaultNamespace,
    pathname,
    featureFlags,
  };
}

function mapDispatchToProps(dispatch: ThunkDispatch<IStoreState, null, Action>) {
  return {
    fetchNamespaces: () => dispatch(actions.namespace.fetchNamespaces()),
    createNamespace: (ns: string) => dispatch(actions.namespace.createNamespace(ns)),
    logout: () => dispatch(actions.auth.logout()),
    push: (path: string) => dispatch(push(path)),
    setNamespace: (ns: string) => dispatch(actions.namespace.setNamespace(ns)),
    getNamespace: (ns: string) => dispatch(actions.namespace.getNamespace(ns)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(Header);
