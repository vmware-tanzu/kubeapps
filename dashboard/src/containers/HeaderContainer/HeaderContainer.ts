import { connect } from "react-redux";
import { RouteComponentProps } from "react-router";
import { push } from "react-router-redux";
import { Dispatch } from "redux";

import actions from "../../actions";
import Header from "../../components/Header";
import { IStoreState } from "../../shared/types";

interface IState extends IStoreState {
  router: RouteComponentProps<{}>;
}

function mapStateToProps({
  auth: { authenticated },
  namespace,
  router: { location: { pathname } },
}: IState) {
  return {
    authenticated,
    namespace,
    pathname,
  };
}

function mapDispatchToProps(dispatch: Dispatch<IStoreState>) {
  return {
    fetchNamespaces: () => dispatch(actions.namespace.fetchNamespaces()),
    logout: (token: string) => dispatch(actions.auth.logout()),
    push: (path: string) => dispatch(push(path)),
    setNamespace: (ns: string) => dispatch(actions.namespace.setNamespace(ns)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(Header);
