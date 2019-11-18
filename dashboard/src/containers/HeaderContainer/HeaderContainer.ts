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
  auth: { authenticated, oidcAuthenticated, defaultNamespace },
  namespace,
  router: {
    location: { pathname },
  },
}: IState) {
  return {
    authenticated,
    namespace,
    defaultNamespace,
    pathname,
    // TODO(mnelson): Allow logoutURL to be configurable.
    logoutUrl: oidcAuthenticated ? "/oauth2/sign_out" : "",
  };
}

function mapDispatchToProps(dispatch: ThunkDispatch<IStoreState, null, Action>) {
  return {
    fetchNamespaces: () => dispatch(actions.namespace.fetchNamespaces()),
    logout: () => dispatch(actions.auth.logout()),
    push: (path: string) => dispatch(push(path)),
    setNamespace: (ns: string) => dispatch(actions.namespace.setNamespace(ns)),
  };
}

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(Header);
