import * as qs from "qs";
import { connect } from "react-redux";
import { RouteComponentProps } from "react-router";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";

import actions from "../../actions";
import OperatorList from "../../components/OperatorList";
import { IStoreState } from "../../shared/types";

function mapStateToProps(
  { operators, clusters: { currentCluster, clusters } }: IStoreState,
  { location }: RouteComponentProps<{}>,
) {
  return {
    cluster: currentCluster,
    namespace: clusters[currentCluster].currentNamespace,
    isFetching: operators.isFetching,
    isOLMInstalled: operators.isOLMInstalled,
    operators: operators.operators,
    error: operators.errors.operator.fetch,
    csvs: operators.csvs,
    filter: qs.parse(location.search, { ignoreQueryPrefix: true }).q || "",
  };
}

function mapDispatchToProps(dispatch: ThunkDispatch<IStoreState, null, Action>) {
  return {
    checkOLMInstalled: (namespace: string) =>
      dispatch(actions.operators.checkOLMInstalled(namespace)),
    getOperators: (namespace: string) => dispatch(actions.operators.getOperators(namespace)),
    getCSVs: (namespace: string) => dispatch(actions.operators.getCSVs(namespace)),
    pushSearchFilter: (filter: string) => dispatch(actions.shared.pushSearchFilter(filter)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(OperatorList);
