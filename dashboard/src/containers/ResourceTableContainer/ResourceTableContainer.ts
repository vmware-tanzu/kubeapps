import { connect } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import ResourceRef from "shared/ResourceRef";
import { IStoreState } from "shared/types";
import actions from "../../actions";
import ResourceTable from "../../components/AppView/ResourceTable/ResourceTable";

function mapStateToProps({ kube }: IStoreState) {
  return {
    resources: kube.items,
  };
}

function mapDispatchToProps(dispatch: ThunkDispatch<IStoreState, null, Action>) {
  return {
    watchResource: (ref: ResourceRef) => dispatch(actions.kube.getAndWatchResource(ref)),
    closeWatch: (ref: ResourceRef) => dispatch(actions.kube.closeWatchResource(ref)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(ResourceTable);
