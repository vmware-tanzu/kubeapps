import { connect } from "react-redux";
import { withRouter } from "react-router";

import { IStoreState } from "../../shared/types";
import Routes from "./Routes";

function mapStateToProps({ auth, clusters: { currentCluster, clusters }, config }: IStoreState) {
  return {
    cluster: currentCluster,
    currentNamespace: clusters[currentCluster].currentNamespace,
    authenticated: auth.authenticated,
  };
}

export default withRouter(connect(mapStateToProps)(Routes));
