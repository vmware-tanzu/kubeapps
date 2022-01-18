// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { push } from "connected-react-router";
import { connect } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import { IStoreState } from "shared/types";
import actions from "../../actions";
import OperatorInstanceUpdateForm from "../../components/OperatorInstanceUpdateForm";

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
  { operators, clusters: { currentCluster, clusters }, config }: IStoreState,
  { match: { params } }: IRouteProps,
) {
  return {
    cluster: currentCluster,
    namespace: clusters[currentCluster].currentNamespace,
    kubeappsCluster: config.kubeappsCluster,
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
    getResource: (
      cluster: string,
      namespace: string,
      csvName: string,
      crdName: string,
      resourceName: string,
    ) =>
      dispatch(actions.operators.getResource(cluster, namespace, csvName, crdName, resourceName)),
    updateResource: (
      cluster: string,
      namespace: string,
      apiVersion: string,
      resource: string,
      resourceName: string,
      body: object,
    ) =>
      dispatch(
        actions.operators.updateResource(
          cluster,
          namespace,
          apiVersion,
          resource,
          resourceName,
          body,
        ),
      ),
    push: (location: string) => dispatch(push(location)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(OperatorInstanceUpdateForm);
