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

function mapStateToProps({ apps }: IStoreState, props: IButtonProps) {
  return {
    namespace: props.namespace,
    releaseName: props.releaseName,
    revision: props.revision,
    loading: apps.isFetching,
  };
}

function mapDispatchToProps(dispatch: ThunkDispatch<IStoreState, null, Action>) {
  return {
    rollbackApp: (namespace: string, releaseName: string, revision: number) =>
      dispatch(actions.apps.rollbackApp(namespace, releaseName, revision)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(RollbackButton);
