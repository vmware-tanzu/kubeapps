import { connect } from "react-redux";
import { withRouter } from "react-router";

import { IStoreState } from "../../shared/types";
import Routes from "./Routes";

function mapStateToProps({ auth, clusters: { currentCluster, clusters }, config }: IStoreState) {
  return {
    cluster: currentCluster,
    namespace: clusters[currentCluster].currentNamespace || auth.defaultNamespace,
    authenticated: auth.authenticated,
    featureFlags: config.featureFlags,
  };
}

export default withRouter(connect(mapStateToProps)(Routes));
