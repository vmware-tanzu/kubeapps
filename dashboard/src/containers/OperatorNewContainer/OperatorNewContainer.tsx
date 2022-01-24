// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { IOperatorNewProps } from "components/OperatorNew/OperatorNew";
import { push } from "connected-react-router";
import { connect } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import { IStoreState } from "shared/types";
import actions from "../../actions";
import OperatorNew from "../../components/OperatorNew";

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
    errors: operators.errors.operator,
    operatorName: params.operator,
  } as Partial<IOperatorNewProps>;
}

function mapDispatchToProps(dispatch: ThunkDispatch<IStoreState, null, Action>) {
  return {
    getOperator: (cluster: string, namespace: string, operatorName: string) =>
      dispatch(actions.operators.getOperator(cluster, namespace, operatorName)),
    createOperator: (
      cluster: string,
      namespace: string,
      name: string,
      channel: string,
      installPlanApproval: string,
      csv: string,
    ) =>
      dispatch(
        actions.operators.createOperator(
          cluster,
          namespace,
          name,
          channel,
          installPlanApproval,
          csv,
        ),
      ),
    push: (location: string) => dispatch(push(location)),
  } as unknown as IOperatorNewProps;
}

export default connect(mapStateToProps, mapDispatchToProps)(OperatorNew);
