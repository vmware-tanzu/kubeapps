import { connect } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";

import actions from "../../actions";
import ApplicationStatus from "../../components/ApplicationStatus";
import { hapi } from "../../shared/hapi/release";
import ResourceRef from "../../shared/ResourceRef";
import { IStoreState } from "../../shared/types";
import { filterByResourceRefs } from "../helpers";

interface IApplicationStatusContainerProps {
  deployRefs: ResourceRef[];
  statefulsetRefs: ResourceRef[];
  daemonsetRefs: ResourceRef[];
  info?: hapi.release.IInfo;
}

function mapStateToProps({ kube }: IStoreState, props: IApplicationStatusContainerProps) {
  const { deployRefs, statefulsetRefs, daemonsetRefs, info } = props;
  return {
    deployments: filterByResourceRefs(deployRefs, kube.items),
    statefulsets: filterByResourceRefs(statefulsetRefs, kube.items),
    daemonsets: filterByResourceRefs(daemonsetRefs, kube.items),
    info,
  };
}

function mapDispatchToProps(
  dispatch: ThunkDispatch<IStoreState, null, Action>,
  props: IApplicationStatusContainerProps,
) {
  return {
    watchWorkloads: () => {
      [...props.deployRefs, ...props.statefulsetRefs, ...props.daemonsetRefs].forEach(r => {
        dispatch(actions.kube.getAndWatchResource(r));
      });
    },
    closeWatches: () => {
      [...props.deployRefs, ...props.statefulsetRefs, ...props.daemonsetRefs].forEach(r => {
        dispatch(actions.kube.closeWatchResource(r));
      });
    },
  };
}

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(ApplicationStatus);
