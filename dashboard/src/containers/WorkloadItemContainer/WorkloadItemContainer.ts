import * as _ from "lodash";
import { connect } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";

import actions from "../../actions";
import WorkloadTableItem from "../../components/AppView/WorkloadTable/WorkloadTableItem";
import ResourceRef from "../../shared/ResourceRef";
import { IStoreState } from "../../shared/types";

interface IWorkloadTableItemContainerProps {
  resourceRef: ResourceRef;
  statusFields: string[];
}

function mapStateToProps({ kube }: IStoreState, props: IWorkloadTableItemContainerProps) {
  const { resourceRef, statusFields } = props;
  return {
    name: resourceRef.name,
    resource: kube.items[resourceRef.getResourceURL()],
    statusFields,
  };
}

function mapDispatchToProps(
  dispatch: ThunkDispatch<IStoreState, null, Action>,
  props: IWorkloadTableItemContainerProps,
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
)(WorkloadTableItem);
