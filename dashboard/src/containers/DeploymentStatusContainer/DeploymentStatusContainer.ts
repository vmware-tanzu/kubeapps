import * as _ from "lodash";
import { connect } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";

import actions from "../../actions";
import DeploymentStatus from "../../components/DeploymentStatus";
import { hapi } from "../../shared/hapi/release";
import ResourceRef from "../../shared/ResourceRef";
import { IStoreState } from "../../shared/types";
import { filterByResourceRefs } from "../helpers";

interface IDeploymentStatusContainerProps {
  deployRefs: ResourceRef[];
  info?: hapi.release.IInfo;
}

function mapStateToProps({ kube }: IStoreState, props: IDeploymentStatusContainerProps) {
  const { deployRefs, info } = props;
  return {
    deployments: filterByResourceRefs(deployRefs, kube.items),
    info,
  };
}

function mapDispatchToProps(
  dispatch: ThunkDispatch<IStoreState, null, Action>,
  props: IDeploymentStatusContainerProps,
) {
  return {
    watchDeployments: () => {
      props.deployRefs.forEach(r => {
        dispatch(actions.kube.getAndWatchResource(r));
      });
    },
    closeWatches: () => {
      props.deployRefs.forEach(r => {
        dispatch(actions.kube.closeWatchResource(r));
      });
    },
  };
}

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(DeploymentStatus);
