// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { push } from "connected-react-router";
import { connect } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import { IStoreState } from "shared/types";
import actions from "../../actions";
import OperatorInstanceForm from "../../components/OperatorInstanceForm";

interface IRouteProps {
  match: {
    params: {
      csv: string;
      crd: string;
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
    getCSV: (cluster: string, namespace: string, name: string) =>
      dispatch(actions.operators.getCSV(cluster, namespace, name)),
    createResource: (
      cluster: string,
      namespace: string,
      apiVersion: string,
      resource: string,
      body: object,
    ) => dispatch(actions.operators.createResource(cluster, namespace, apiVersion, resource, body)),
    push: (location: string) => dispatch(push(location)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(OperatorInstanceForm);
