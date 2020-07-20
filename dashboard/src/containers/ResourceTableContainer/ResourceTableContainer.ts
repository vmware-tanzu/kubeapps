import { connect } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";

import actions from "../../actions";
import ResourceTable from "../../components/AppView/ResourceTable/ResourceTable.v2";
import ResourceRef from "../../shared/ResourceRef";
import { IStoreState } from "../../shared/types";

interface IResourceTableItemContainerProps {
  title: string;
  resourceRefs: ResourceRef[];
}

function mapStateToProps({ kube }: IStoreState, props: IResourceTableItemContainerProps) {
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
