import { CdsIcon } from "@cds/react/icon";
import actions from "actions";
import ConfirmDialog from "components/ConfirmDialog/ConfirmDialog";
import { push } from "connected-react-router";
import { useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import { IStoreState } from "shared/types";
import { app } from "shared/url";
import { hapi } from "../../../../shared/hapi/release";
import StatusAwareButton from "../StatusAwareButton";

interface IDeleteButtonProps {
  cluster: string;
  namespace: string;
  releaseName: string;
  releaseStatus: hapi.release.IStatus | undefined | null;
}

export default function DeleteButton({
  cluster,
  namespace,
  releaseName,
  releaseStatus,
}: IDeleteButtonProps) {
  const [modalIsOpen, setModal] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const dispatch: ThunkDispatch<IStoreState, null, Action> = useDispatch();
  const error = useSelector((state: IStoreState) => state.apps.error);

  const openModal = () => setModal(true);
  const closeModal = () => setModal(false);
  const handleDeleteClick = async () => {
    setDeleting(true);
    // Purge the release in any case
    const deleted = await dispatch(actions.apps.deleteApp(cluster, namespace, releaseName, true));
    setDeleting(false);
    if (deleted) {
      dispatch(push(app.apps.list(cluster, namespace)));
    }
  };

  return (
    <>
      <StatusAwareButton
        status="danger"
        onClick={openModal}
        releaseStatus={releaseStatus}
        id="delete-button"
      >
        <CdsIcon shape="trash" /> Delete
      </StatusAwareButton>
      <ConfirmDialog
        modalIsOpen={modalIsOpen}
        loading={deleting}
        onConfirm={handleDeleteClick}
        closeModal={closeModal}
        headerText={"Delete application"}
        confirmationText="Are you sure you want to delete the application?"
        error={error}
      />
    </>
  );
}
