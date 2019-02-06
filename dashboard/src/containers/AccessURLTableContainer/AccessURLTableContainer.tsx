import * as _ from "lodash";
import { connect } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";

import actions from "../../actions";
import AccessURLTable from "../../components/AppView/AccessURLTable";
import ResourceRef from "../../shared/ResourceRef";
import { IStoreState } from "../../shared/types";
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
  };
}

function mapDispatchToProps(
  dispatch: ThunkDispatch<IStoreState, null, Action>,
  props: IAccessURLTableContainerProps,
) {
  return {
    // Fetch each Ingress in the ingressRefs. We don't have an action for
    // fetching Services as they are assumed to be fetched and watched by the
    // ServiceItemContainers.
    fetchIngresses: () => {
      props.ingressRefs.forEach(r => {
        dispatch(actions.kube.getResource(r));
      });
    },
    watchServices: () => {
      props.serviceRefs.forEach(r => {
        dispatch(actions.kube.getAndWatchResource(r));
      });
    },
    closeWatches: () => {
      props.serviceRefs.forEach(r => {
        dispatch(actions.kube.closeWatchResource(r));
      });
    },
  };
}

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(AccessURLTable);
