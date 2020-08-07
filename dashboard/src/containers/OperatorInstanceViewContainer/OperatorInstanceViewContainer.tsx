import { push } from "connected-react-router";
import { connect } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";

import actions from "../../actions";
import OperatorInstance from "../../components/OperatorInstance";
import { IResource, IStoreState } from "../../shared/types";

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
  { apps, clusters: { currentCluster, clusters }, operators }: IStoreState,
  { match: { params } }: IRouteProps,
) {
  return {
    cluster: currentCluster,
    namespace: clusters[currentCluster].currentNamespace,
    csvName: params.csv,
    crdName: params.crd,
    instanceName: params.instanceName,
    isFetching: operators.isFetching,
    resource: operators.resource,
    csv: operators.csv,
    errors: operators.errors.resource,
  };
}

function mapDispatchToProps(dispatch: ThunkDispatch<IStoreState, null, Action>) {
  return {
    getResource: (namespace: string, csvName: string, crdName: string, resourceName: string) =>
      dispatch(actions.operators.getResource(namespace, csvName, crdName, resourceName)),
    deleteResource: (namespace: string, crdName: string, resource: IResource) =>
      dispatch(actions.operators.deleteResource(namespace, crdName, resource)),
    push: (location: string) => dispatch(push(location)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(OperatorInstance);
