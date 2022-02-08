// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { push } from "connected-react-router";
import { connect } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import { IResource, IStoreState } from "shared/types";
import actions from "../../actions";
import OperatorInstance from "../../components/OperatorInstance";

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
  { clusters: { currentCluster, clusters }, config, operators }: IStoreState,
  { match: { params } }: IRouteProps,
) {
  return {
    cluster: currentCluster,
    namespace: clusters[currentCluster].currentNamespace,
    kubeappsCluster: config.kubeappsCluster,
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
    getResource: (
      cluster: string,
      namespace: string,
      csvName: string,
      crdName: string,
      resourceName: string,
    ) =>
      dispatch(actions.operators.getResource(cluster, namespace, csvName, crdName, resourceName)),
    deleteResource: (cluster: string, namespace: string, crdName: string, resource: IResource) =>
      dispatch(actions.operators.deleteResource(cluster, namespace, crdName, resource)),
    push: (location: string) => dispatch(push(location)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(OperatorInstance);
