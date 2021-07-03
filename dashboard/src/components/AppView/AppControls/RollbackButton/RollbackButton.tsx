import { CdsIcon } from "@cds/react/icon";
import actions from "actions";
import { useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import { IStoreState } from "shared/types";
import RollbackDialog from "./RollbackDialog";
import { hapi } from "../../../../shared/hapi/release";
import StatusAwareButton from "../StatusAwareButton";

export interface IRollbackButtonProps {
  cluster: string;
  namespace: string;
  releaseName: string;
  revision: number;
  releaseStatus: hapi.release.IStatus | undefined | null;
}

function RollbackButton({
  cluster,
  namespace,
  releaseName,
  revision,
  releaseStatus,
}: IRollbackButtonProps) {
  const [modalIsOpen, setModalOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const dispatch: ThunkDispatch<IStoreState, null, Action> = useDispatch();
  const error = useSelector((state: IStoreState) => state.apps.error);
  const openModal = () => setModalOpen(true);
  const closeModal = () => setModalOpen(false);
  const handleRollback = async (r: number) => {
    setLoading(true);
    const success = await dispatch(actions.apps.rollbackApp(cluster, namespace, releaseName, r));
    setLoading(false);
    if (success) {
      closeModal();
    }
  };
  return (
    <>
      <RollbackDialog
        modalIsOpen={modalIsOpen}
        onConfirm={handleRollback}
        loading={loading}
        closeModal={closeModal}
        currentRevision={revision}
        error={error}
      />
      <StatusAwareButton
        status="primary"
        onClick={openModal}
        releaseStatus={releaseStatus}
        id="rollback-button"
      >
        <CdsIcon shape="rewind" /> Rollback
      </StatusAwareButton>
    </>
  );
}

export default RollbackButton;
