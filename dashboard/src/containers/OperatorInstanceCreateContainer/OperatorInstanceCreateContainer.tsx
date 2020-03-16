import { push } from "connected-react-router";
import { connect } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";

import OperatorInstanceForm from "components/OperatorInstanceForm";
import actions from "../../actions";
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
  { operators, namespace }: IStoreState,
  { match: { params } }: IRouteProps,
) {
  return {
    namespace: namespace.current,
    isFetching: operators.isFetching,
    csv: operators.csv,
    errors: operators.errors,
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
