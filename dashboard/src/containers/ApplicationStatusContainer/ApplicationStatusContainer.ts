import { InstalledPackageDetail } from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { connect } from "react-redux";
import ApplicationStatus from "../../components/ApplicationStatus";
import { ResourceRef } from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { IStoreState } from "../../shared/types";
import { filterByResourceRefs } from "../helpers";

interface IApplicationStatusContainerProps {
  deployRefs: ResourceRef[];
  statefulsetRefs: ResourceRef[];
  daemonsetRefs: ResourceRef[];
  info?: InstalledPackageDetail;
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

export default connect(mapStateToProps)(ApplicationStatus);
