import * as _ from "lodash";
import { connect } from "react-redux";

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

export default connect(mapStateToProps)(DeploymentStatus);
