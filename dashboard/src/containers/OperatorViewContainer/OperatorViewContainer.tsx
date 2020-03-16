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
  { operators, namespace }: IStoreState,
  { match: { params } }: IRouteProps,
) {
  return {
    namespace: namespace.current,
    isFetching: operators.isFetching,
    operator: operators.operator,
    error: operators.errors.fetch,
    operatorName: params.operator,
  };
}

function mapDispatchToProps(dispatch: ThunkDispatch<IStoreState, null, Action>) {
  return {
    getOperator: (namespace: string, operatorName: string) =>
      dispatch(actions.operators.getOperator(namespace, operatorName)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(OperatorView);
