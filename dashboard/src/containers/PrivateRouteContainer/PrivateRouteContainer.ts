import { connect } from "react-redux";
import { withRouter } from "react-router";

import PrivateRoute from "../../components/PrivateRoute";
import { IStoreState } from "../../shared/types";

function mapStateToProps({
  auth: { authenticated, oidcAuthenticated, sessionExpired },
  clusters: { currentCluster },
}: IStoreState) {
  return {
    sessionExpired,
    authenticated,
    oidcAuthenticated,
    cluster: currentCluster,
  };
}

export default withRouter(connect(mapStateToProps)(PrivateRoute));
