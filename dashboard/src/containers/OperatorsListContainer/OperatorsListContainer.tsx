// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import qs from "qs";
import { connect } from "react-redux";
import { RouteComponentProps } from "react-router-dom";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import { IStoreState } from "shared/types";
import actions from "../../actions";
import OperatorList from "../../components/OperatorList";

function mapStateToProps(
  { operators, clusters: { currentCluster, clusters }, config }: IStoreState,
  { location }: RouteComponentProps<{}>,
) {
  return {
    cluster: currentCluster,
    namespace: clusters[currentCluster].currentNamespace,
    kubeappsCluster: config.kubeappsCluster,
    isFetching: operators.isFetching,
    isOLMInstalled: operators.isOLMInstalled,
    operators: operators.operators,
    error: operators.errors.operator.fetch,
    csvs: operators.csvs,
    filter: qs.parse(location.search, { ignoreQueryPrefix: true }),
  };
}

function mapDispatchToProps(dispatch: ThunkDispatch<IStoreState, null, Action>) {
  return {
    checkOLMInstalled: (cluster: string, namespace: string) =>
      dispatch(actions.operators.checkOLMInstalled(cluster, namespace)),
    getOperators: (cluster: string, namespace: string) =>
      dispatch(actions.operators.getOperators(cluster, namespace)),
    getCSVs: (cluster: string, namespace: string) =>
      dispatch(actions.operators.getCSVs(cluster, namespace)),
    pushSearchFilter: (filter: string) => dispatch(actions.shared.pushSearchFilter(filter)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(OperatorList);
