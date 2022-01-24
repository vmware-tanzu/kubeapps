import { connect } from "react-redux";
import { withRouter } from "react-router-dom";
import { IStoreState } from "shared/types";
import PrivateRoute from "../../components/PrivateRoute";

function mapStateToProps({
  auth: { authenticated, oidcAuthenticated, sessionExpired },
}: IStoreState) {
  return {
    sessionExpired,
    authenticated,
    oidcAuthenticated,
  };
}

export default withRouter(connect(mapStateToProps)(PrivateRoute));
