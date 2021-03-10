import { CdsButton } from "@cds/react/button";
import { CdsIcon } from "@cds/react/icon";
import actions from "actions";
import { useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import { IStoreState } from "shared/types";
import RollbackDialog from "./RollbackDialog";

export interface IRollbackButtonProps {
  cluster: string;
  namespace: string;
  releaseName: string;
  revision: number;
}

function RollbackButton({ cluster, namespace, releaseName, revision }: IRollbackButtonProps) {
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
      <CdsButton status="primary" onClick={openModal}>
        <CdsIcon shape="rewind" /> Rollback
      </CdsButton>
    </>
  );
}

export default RollbackButton;
