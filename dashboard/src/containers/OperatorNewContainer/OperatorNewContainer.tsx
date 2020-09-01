import { push } from "connected-react-router";
import { connect } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";

import actions from "../../actions";
import OperatorNew from "../../components/OperatorNew";
import { IStoreState } from "../../shared/types";

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
    UI: config.featureFlags.ui,
  };
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
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(OperatorNew);
