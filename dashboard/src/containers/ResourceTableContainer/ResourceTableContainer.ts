import { connect } from "react-redux";
import { IStoreState } from "shared/types";
import ResourceTable from "../../components/AppView/ResourceTable/ResourceTable";

// TODO(minelson): This container is no longer needed... switch to useSelector
// in component.
function mapStateToProps({ kube }: IStoreState) {
  return {
    resources: kube.items,
  };
}

export default connect(mapStateToProps)(ResourceTable);
