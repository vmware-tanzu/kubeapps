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
  { operators, namespace }: IStoreState,
  { match: { params } }: IRouteProps,
) {
  return {
    namespace: namespace.current,
    isFetching: operators.isFetching,
    operator: operators.operator,
    errors: operators.errors.operator,
    operatorName: params.operator,
  };
}

function mapDispatchToProps(dispatch: ThunkDispatch<IStoreState, null, Action>) {
  return {
    getOperator: (namespace: string, operatorName: string) =>
      dispatch(actions.operators.getOperator(namespace, operatorName)),
    createOperator: (
      namespace: string,
      name: string,
      channel: string,
      installPlanApproval: string,
      csv: string,
    ) =>
      dispatch(
        actions.operators.createOperator(namespace, name, channel, installPlanApproval, csv),
      ),
    push: (location: string) => dispatch(push(location)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(OperatorNew);
