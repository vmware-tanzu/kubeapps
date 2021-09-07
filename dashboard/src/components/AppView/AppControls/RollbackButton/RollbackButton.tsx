import { CdsIcon } from "@cds/react/icon";
import actions from "actions";
import { InstalledPackageStatus } from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import { useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import { IStoreState } from "shared/types";
import StatusAwareButton from "../StatusAwareButton/StatusAwareButton";
import RollbackDialog from "./RollbackDialog";

export interface IRollbackButtonProps {
  cluster: string;
  namespace: string;
  releaseName: string;
  revision: number;
  releaseStatus: InstalledPackageStatus | undefined | null;
  plugin: Plugin;
}

function RollbackButton({
  cluster,
  namespace,
  releaseName,
  revision,
  releaseStatus,
  plugin,
}: IRollbackButtonProps) {
  const [modalIsOpen, setModalOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const dispatch: ThunkDispatch<IStoreState, null, Action> = useDispatch();
  const error = useSelector((state: IStoreState) => state.apps.error);
  const openModal = () => setModalOpen(true);
  const closeModal = () => setModalOpen(false);
  const handleRollback = async (r: number) => {
    setLoading(true);
    const success = await dispatch(
      actions.apps.rollbackApp(cluster, namespace, releaseName, r, plugin),
    );
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
