import { push } from "connected-react-router";
import { connect } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";

import actions from "../../actions";
import OperatorInstanceForm from "../../components/OperatorInstanceForm";
import { IStoreState } from "../../shared/types";

interface IRouteProps {
  match: {
    params: {
      csv: string;
      crd: string;
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
    errors: {
      fetch: operators.errors.csv.fetch,
      create: operators.errors.resource.create,
    },
    csvName: params.csv,
    crdName: params.crd,
  };
}

function mapDispatchToProps(dispatch: ThunkDispatch<IStoreState, null, Action>) {
  return {
    getCSV: (namespace: string, name: string) =>
      dispatch(actions.operators.getCSV(namespace, name)),
    createResource: (namespace: string, apiVersion: string, resource: string, body: object) =>
      dispatch(actions.operators.createResource(namespace, apiVersion, resource, body)),
    push: (location: string) => dispatch(push(location)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(OperatorInstanceForm);
