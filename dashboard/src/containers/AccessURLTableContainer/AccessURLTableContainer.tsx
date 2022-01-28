// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { connect } from "react-redux";
import { ResourceRef } from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { IStoreState } from "shared/types";
import AccessURLTable from "../../components/AppView/AccessURLTable";
import { filterByResourceRefs } from "../helpers";

interface IAccessURLTableContainerProps {
  serviceRefs: ResourceRef[];
  ingressRefs: ResourceRef[];
}

function mapStateToProps({ kube }: IStoreState, props: IAccessURLTableContainerProps) {
  // Extract the Services and Ingresses form the Redux state using the keys for
  // each ResourceRef.
  return {
    services: filterByResourceRefs(props.serviceRefs, kube.items),
    ingresses: filterByResourceRefs(props.ingressRefs, kube.items),
    ingressRefs: props.ingressRefs,
  };
}

export default connect(mapStateToProps)(AccessURLTable);
