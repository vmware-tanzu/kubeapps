import { push } from "connected-react-router";
import { connect } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";

import actions from "../../actions";
import OperatorView from "../../components/OperatorView";
import { IStoreState } from "../../shared/types";

interface IRouteProps {
  match: {
    params: {
      operator: string;
    };
  };
}

function mapStateToProps(
  { operators, clusters: { currentCluster, clusters } }: IStoreState,
  { match: { params } }: IRouteProps,
) {
  return {
    cluster: currentCluster,
    namespace: clusters[currentCluster].currentNamespace,
    isFetching: operators.isFetching,
    operator: operators.operator,
    error: operators.errors.operator.fetch,
    operatorName: params.operator,
    csv: operators.csv,
  };
}

function mapDispatchToProps(dispatch: ThunkDispatch<IStoreState, null, Action>) {
  return {
    getOperator: (namespace: string, operatorName: string) =>
      dispatch(actions.operators.getOperator(namespace, operatorName)),
    push: (location: string) => dispatch(push(location)),
    getCSV: (namespace: string, name: string) =>
      dispatch(actions.operators.getCSV(namespace, name)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(OperatorView);
