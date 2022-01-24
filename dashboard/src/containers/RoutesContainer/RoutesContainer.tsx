// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { connect } from "react-redux";
import { withRouter } from "react-router-dom";
import { IStoreState } from "shared/types";
import Routes from "./Routes";

function mapStateToProps({ auth, clusters: { currentCluster, clusters }, config }: IStoreState) {
  return {
    cluster: currentCluster,
    currentNamespace: clusters[currentCluster].currentNamespace,
    authenticated: auth.authenticated,
    featureFlags: config.featureFlags,
  };
}

export default withRouter(connect(mapStateToProps)(Routes));
