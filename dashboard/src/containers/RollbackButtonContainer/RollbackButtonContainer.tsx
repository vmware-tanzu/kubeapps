import { connect } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";

import actions from "../../actions";
import RollbackButton from "../../components/AppView/AppControls/RollbackButton";
import { IStoreState } from "../../shared/types";

interface IButtonProps {
  namespace: string;
  releaseName: string;
  revision: number;
}

function mapStateToProps({ apps, clusters }: IStoreState, props: IButtonProps) {
  return {
    cluster: clusters.currentCluster,
    namespace: props.namespace,
    releaseName: props.releaseName,
    revision: props.revision,
    loading: apps.isFetching,
  };
}

function mapDispatchToProps(dispatch: ThunkDispatch<IStoreState, null, Action>) {
  return {
    rollbackApp: (cluster: string, namespace: string, releaseName: string, revision: number) =>
      dispatch(actions.apps.rollbackApp(cluster, namespace, releaseName, revision)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(RollbackButton);
