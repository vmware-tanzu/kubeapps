import { connect } from "react-redux";
import { RouteComponentProps } from "react-router";
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
    logout: (token: string) => dispatch(actions.auth.logout()),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(Header);
