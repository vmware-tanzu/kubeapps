import { connect } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";

import actions from "../../actions";
import OperatorInstance from "../../components/OperatorInstance";
import { IStoreState } from "../../shared/types";

interface IRouteProps {
  match: {
    params: {
      csv: string;
      crd: string;
      instanceName: string;
    };
  };
}
function mapStateToProps(
  { apps, namespace, operators }: IStoreState,
  { match: { params } }: IRouteProps,
) {
  return {
    namespace: namespace.current,
    csvName: params.csv,
    crdName: params.crd,
    instanceName: params.instanceName,
    isFetching: operators.isFetching,
    resource: operators.resource,
    csv: operators.csv,
    error: operators.errors.fetch,
  };
}

function mapDispatchToProps(dispatch: ThunkDispatch<IStoreState, null, Action>) {
  return {
    getResource: (namespace: string, csvName: string, crdName: string, resourceName: string) =>
      dispatch(actions.operators.getResource(namespace, csvName, crdName, resourceName)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(OperatorInstance);
