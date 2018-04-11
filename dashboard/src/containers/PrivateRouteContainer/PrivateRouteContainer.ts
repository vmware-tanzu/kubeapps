import { connect } from "react-redux";
import { withRouter } from "react-router";

import PrivateRoute from "../../components/PrivateRoute";
import { IStoreState } from "../../shared/types";

function mapStateToProps({ auth: { authenticated } }: IStoreState) {
  return {
    authenticated,
  };
}

export default withRouter(connect(mapStateToProps)(PrivateRoute));
