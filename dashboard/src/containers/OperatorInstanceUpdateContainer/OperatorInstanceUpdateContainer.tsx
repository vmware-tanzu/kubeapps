import { push } from "connected-react-router";
import { connect } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";

import actions from "../../actions";
import OperatorInstanceUpdateForm from "../../components/OperatorInstanceUpdateForm";
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
  { operators, clusters: { currentCluster, clusters } }: IStoreState,
  { match: { params } }: IRouteProps,
) {
  return {
    cluster: currentCluster,
    namespace: clusters[currentCluster].currentNamespace,
    isFetching: operators.isFetching,
    csv: operators.csv,
    errors: operators.errors.resource,
    csvName: params.csv,
    crdName: params.crd,
    resourceName: params.instanceName,
    resource: operators.resource,
  };
}

function mapDispatchToProps(dispatch: ThunkDispatch<IStoreState, null, Action>) {
  return {
    getResource: (namespace: string, csvName: string, crdName: string, resourceName: string) =>
      dispatch(actions.operators.getResource(namespace, csvName, crdName, resourceName)),
    updateResource: (
      namespace: string,
      apiVersion: string,
      resource: string,
      resourceName: string,
      body: object,
    ) =>
      dispatch(
        actions.operators.updateResource(namespace, apiVersion, resource, resourceName, body),
      ),
    push: (location: string) => dispatch(push(location)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(OperatorInstanceUpdateForm);
