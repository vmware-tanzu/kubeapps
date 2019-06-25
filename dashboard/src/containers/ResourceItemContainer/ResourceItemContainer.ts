import { connect } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";

import actions from "../../actions";
import ResourceTableItem from "../../components/AppView/ResourceTable/ResourceItem/ResourceTableItem";
import ResourceRef from "../../shared/ResourceRef";
import { IStoreState } from "../../shared/types";

interface IResourceTableItemContainerProps {
  resourceRef: ResourceRef;
}

function mapStateToProps({ kube }: IStoreState, props: IResourceTableItemContainerProps) {
  const { resourceRef } = props;
  return {
    name: resourceRef.name,
    resource: kube.items[resourceRef.getResourceURL()],
  };
}

function mapDispatchToProps(
  dispatch: ThunkDispatch<IStoreState, null, Action>,
  props: IResourceTableItemContainerProps,
) {
  const { resourceRef } = props;
  return {
    watchResource: () => dispatch(actions.kube.getAndWatchResource(resourceRef)),
    closeWatch: () => dispatch(actions.kube.closeWatchResource(resourceRef)),
  };
}

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(ResourceTableItem);
