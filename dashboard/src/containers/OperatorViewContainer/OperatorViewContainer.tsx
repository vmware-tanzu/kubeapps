// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { push } from "connected-react-router";
import { connect } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import { IStoreState } from "shared/types";
import actions from "../../actions";
import OperatorView from "../../components/OperatorView";

interface IRouteProps {
  match: {
    params: {
      operator: string;
    };
  };
}

function mapStateToProps(
  { operators, clusters: { currentCluster, clusters }, config }: IStoreState,
  { match: { params } }: IRouteProps,
) {
  return {
    cluster: currentCluster,
    namespace: clusters[currentCluster].currentNamespace,
    kubeappsCluster: config.kubeappsCluster,
    isFetching: operators.isFetching,
    operator: operators.operator,
    error: operators.errors.operator.fetch,
    operatorName: params.operator,
    csv: operators.csv,
  };
}

function mapDispatchToProps(dispatch: ThunkDispatch<IStoreState, null, Action>) {
  return {
    getOperator: (cluster: string, namespace: string, operatorName: string) =>
      dispatch(actions.operators.getOperator(cluster, namespace, operatorName)),
    push: (location: string) => dispatch(push(location)),
    getCSV: (cluster: string, namespace: string, name: string) =>
      dispatch(actions.operators.getCSV(cluster, namespace, name)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(OperatorView);
